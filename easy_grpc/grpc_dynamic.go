package easy_grpc

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type DynamicGRPC struct {
	registryFiles *protoregistry.Files
	protoFiles    []string
	path          string
}

const (
	TEMP_DIR    = "descriptor.tmp"
	PB_FILENAME = "descriptor.pb"
)

func GenerateDynamicGRPC(path string) (*DynamicGRPC, error) {
	tmp, err := os.MkdirTemp(path, TEMP_DIR)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	tmpAbs, err := filepath.Abs(tmp)
	if err != nil {
		return nil, err
	}

	descripterFilename := fmt.Sprintf("%s/%s", tmpAbs, PB_FILENAME)

	protoFiles, err := scanProtoFiles(path)
	if err != nil {
		return nil, err
	}

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var protocArgs []string = make([]string, 0)
	if len(protoFiles) > 100 {
		argTempFile, err := os.CreateTemp(path, "protoc")
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(argTempFile.Name(), os.O_RDWR, 0755)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		for _, protoFile := range protoFiles {
			f.WriteString(fmt.Sprintf("%s\n", protoFile))
		}
		argTempFileAbs, err := filepath.Abs(argTempFile.Name())
		if err != nil {
			return nil, err
		}
		protocArgs = append(protocArgs, fmt.Sprintf("--proto_path=%s", pathAbs))
		protocArgs = append(protocArgs, fmt.Sprintf("--descriptor_set_out=%s", descripterFilename))
		protocArgs = append(protocArgs, "--include_imports")
		protocArgs = append(protocArgs, fmt.Sprintf("@%s", argTempFileAbs))
	} else if len(protoFiles) > 0 && len(protoFiles) <= 100 {
		protocArgs = append(protocArgs, fmt.Sprintf("--descriptor_set_out=%s", descripterFilename))
		protocArgs = append(protocArgs, "--include_imports")
		protocArgs = append(protocArgs, fmt.Sprintf("--proto_path=%s", pathAbs))
		protocArgs = append(protocArgs, protoFiles...)
	} else {
		err = errors.New("cannot found proto files")
		return nil, err
	}
	registry, err := createProtoRegistry(descripterFilename, protocArgs...)
	if err != nil {
		return nil, err
	}
	return &DynamicGRPC{
		protoFiles:    protoFiles,
		registryFiles: registry,
		path:          pathAbs,
	}, nil
}

func (d *DynamicGRPC) GetDescriptorFileByName(filename string) (protoreflect.FileDescriptor, error) {
	baseFilename := filepath.Base(filename)
	for _, protoFilename := range d.protoFiles {
		if filepath.Base(protoFilename) == baseFilename {
			p := protoFilename[len(d.path)+1:]
			fmt.Println(p)
			return d.registryFiles.FindFileByPath(p)
		}
	}
	return nil, errors.New("no such proto file regist: " + filename)
}

type DynamicAPI struct {
	service        string
	method         string
	requestMessage *dynamicpb.Message
	replyMessage   *dynamicpb.Message
}

func NewDynamicAPI(pfd protoreflect.FileDescriptor, service string, method string) *DynamicAPI {
	pfServiceDescriptor := pfd.Services().ByName(protoreflect.Name(service))
	pfMethodDescriptor := pfServiceDescriptor.Methods().ByName(protoreflect.Name(method))
	pfRequestMessage := pfMethodDescriptor.Input()
	pfReplyMessage := pfMethodDescriptor.Output()
	dyRequestMessage := dynamicpb.NewMessage(pfRequestMessage)
	dyReplyMessage := dynamicpb.NewMessage(pfReplyMessage)
	return &DynamicAPI{
		service:        service,
		method:         fmt.Sprintf("%s/%s", string(pfServiceDescriptor.FullName()), string(pfMethodDescriptor.Name())),
		requestMessage: dyRequestMessage,
		replyMessage:   dyReplyMessage,
	}
}

func (d *DynamicAPI) Invoke(handler *grpcHandler, jsonMessage string) (string, error) {
	if err := protojson.Unmarshal([]byte(jsonMessage), d.requestMessage); err != nil {
		return "", err
	}
	if err := handler.Conn.Invoke(handler.Ctx, d.method, d.requestMessage, d.replyMessage); err != nil {
		return "", err
	}
	resp, err := protojson.Marshal(d.replyMessage)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func scanProtoFiles(path string) ([]string, error) {
	var protoFiles = make([]string, 0)
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			protoFiles = append(protoFiles, absPath)
		}
		return nil
	})
	return protoFiles, err
}

func createProtoRegistry(descriptorFilename string, protocArgs ...string) (*protoregistry.Files, error) {

	cmd := exec.Command("protoc",
		protocArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	marshalledDescriptorSet, err := os.ReadFile(descriptorFilename)
	if err != nil {
		return nil, err
	}
	descriptorSet := descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(marshalledDescriptorSet, &descriptorSet)
	if err != nil {
		return nil, err
	}
	files, err := protodesc.NewFiles(&descriptorSet)
	if err != nil {
		return nil, err
	}

	return files, nil
}
