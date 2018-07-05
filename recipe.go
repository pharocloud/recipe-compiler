package main

import (
	"archive/tar"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v2"
)

type File struct {
	Name    string
	Type    string
	Content string
}

type Recipe struct {
	Configuration struct {
		Image    string
		Commands []string
		Files    []File `yaml:"write_files"`
	}
	RunTime struct {
		VM      string
		Command string
		Files   []File `yaml:"write_files"`
	}
}

func Parse(data []byte) (Recipe, error) {
	var r Recipe
	err := yaml.Unmarshal([]byte(data), &r)
	return r, err
}

func (r Recipe) WriteTo(w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	for _, f := range r.Configuration.Files {
		err := tarRecipeFile(tw, "conf/", f)
		if err != nil {
			return err
		}
	}

	for _, f := range r.RunTime.Files {
		err := tarRecipeFile(tw, "run/", f)
		if err != nil {
			return err
		}
	}
	d, err := generateDockerfile(r)
	if err != nil {
		return err
	}

	tarContent(tw, "Dockerfile", []byte(d))
	return nil
}

func tarRecipeFile(tw *tar.Writer, prefix string, file File) error {
	return tarContent(tw, prefix+file.Name, []byte(file.Content))
}

func tarContent(tw *tar.Writer, name string, content []byte) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("failed to write header for %d %s: %v", hdr.Size, name, err)
	}

	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("failed to write body of %s: %v", name, err)
	}

	return nil
}

func generateDockerfile(r Recipe) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "FROM %s AS configuration\n", r.Configuration.Image)
	fmt.Fprint(&b, "COPY conf/ /var/pharo/images/default/\n")
	for _, c := range r.Configuration.Commands {
		fmt.Fprintf(&b, "RUN /usr/local/bin/pharo /var/pharo/images/default/Pharo.image %s\n", c)
	}
	fmt.Fprintf(&b, "FROM %s\n", r.RunTime.VM)
	fmt.Fprint(&b, "COPY --from=configuration /var/pharo/images/default/Pharo.{image,changes} /var/pharo/images/default/\n")
	fmt.Fprint(&b, "COPY run/ /var/pharo/images/default/\n")
	fmt.Fprintf(&b, "CMD /usr/local/bin/pharo /var/pharo/images/default/Pharo.image %s\n", r.RunTime.Command)
	return b.String(), nil
}
