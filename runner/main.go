package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func addFile(path string, tw *tar.Writer) {
	payload, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err, " :unable to read file")
	}

	tarHeader := &tar.Header{
		Name: path,
		Size: int64(len(payload)),
	}
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, " :unable to write tar header")
	}
	_, err = tw.Write(payload)
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
	}
}

func build(version string, cli *client.Client) string {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerFile := "Dockerfile"
	readDockerFile := []byte(fmt.Sprintf(`
FROM golang:%s-alpine

WORKDIR payload/
RUN go mod init github.com/example
# RUN go get golang.org/x/crypto/blake2b

ADD payload/ .
RUN go build ./main.go

CMD ./main
`, strings.Trim(version, " ")))

	tarHeader := &tar.Header{
		Name: dockerFile,
		Size: int64(len(readDockerFile)),
	}
	err := tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, " :unable to write tar header")
	}
	_, err = tw.Write(readDockerFile)
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
	}

	//addFile("payload/go.mod", tw)
	//addFile("payload/go.sum", tw)
	addFile("payload/main.go", tw)

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	imageBuildResponse, err := cli.ImageBuild(
		context.Background(),
		dockerFileTarReader,
		types.ImageBuildOptions{
			Context:    dockerFileTarReader,
			Dockerfile: dockerFile,
			Remove:     true})
	if err != nil {
		log.Fatal(err, " :unable to build docker image")
	}
	defer imageBuildResponse.Body.Close()
	/*_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		log.Fatal(err, " :unable to read image build response")
	}*/

	data, err := io.ReadAll(imageBuildResponse.Body)
	if err != nil {
		log.Fatal(err, " :unable to build docker image")
	}
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.Contains(line, "\"ID\":") {
			return strings.Split(strings.Split(line, "sha256:")[1], "\"")[0]
		}
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	panic(fmt.Sprintf("%s", lines))
}

func buildAndRun(version string, cli *client.Client) {
	cli.ContainerRemove(context.Background(), "name", container.RemoveOptions{})

	image := build(version, cli)
	ctr, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: image,
		Tty:   true,
	}, nil, nil, nil, "name")
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStart(context.Background(), ctr.ID, container.StartOptions{})
	if err != nil {
		panic(err)
	}

	logsStream, err := cli.ContainerLogs(context.Background(), ctr.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	logs, err := io.ReadAll(logsStream)
	fmt.Printf("[%s] %s\n", version, logs)
}

func main() {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.46"), client.WithHost("unix:///Users/dean/.docker/run/docker.sock"))
	if err != nil {
		panic(err)
	}

	//buildAndRun("1.18  ", cli)
	//buildAndRun("1.19  ", cli)
	buildAndRun("1.20  ", cli)
	buildAndRun("1.21.0", cli)
	//buildAndRun("1.22  ", cli)
	//buildAndRun("1.23  ", cli)
}
