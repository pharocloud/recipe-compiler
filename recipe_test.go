package main

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"
)

var recipe1 = `
configuration:
  image: pharo/image:61
  commands:
    - "st --save --quit config.st"
  write_files:
    - name: config.st
      content-type: plain/text
      content: |
        ConfigurationOfZincHTTPComponents project latestVersion load: 'WebSocket'.

runtime:
  vm: pharo/vm:61
  command: --no-quit
  write_files:
    - name: startup.st
      content-type: plain/text
      content: |
        ZnServer startDefaultOn: 8080.

        ZnServer default delegate
          map: 'ws-chatroom-client'
          to: [ :request | ZnResponse ok: (ZnEntity html: ZnWebSocketChatroomHandler clientHtml) ];
          map: 'ws-chatroom'
          to: (ZnWebSocketDelegate map: 'ws-chatroom' to: ZnWebSocketChatroomHandler new).
`

var dockerfile = `FROM pharo/image:61 AS configuration
COPY conf/ /var/pharo/images/default/
RUN /usr/local/bin/pharo /var/pharo/images/default/Pharo.image st --save --quit config.st
FROM pharo/vm:61
COPY --from=configuration /var/pharo/images/default/Pharo.{image,changes} /var/pharo/images/default/
COPY run/ /var/pharo/images/default/
CMD /usr/local/bin/pharo /var/pharo/images/default/Pharo.image --no-quit
`

func TestParsing(t *testing.T) {
	r, err := Parse([]byte(recipe1))
	if err != nil {
		t.Fatal(err)
	}
	if r.RunTime.VM != "pharo/vm:61" {
		t.Errorf("runtime.vm expected pharo/vm:61, got: %s", r.RunTime.VM)
	}

	d, err := generateDockerfile(r)
	if err != nil {
		t.Fatal(err)
	}

	if d != dockerfile {
		t.Errorf("Dockerfile did not match, expected:\n%s\ngot:\n%s\n", dockerfile, d)
	}
}

func TestTar(t *testing.T) {
	r, err := Parse([]byte(recipe1))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = r.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		"conf/config.st": false,
		"run/startup.st": false,
		"Dockerfile":     false,
	}

	tr := tar.NewReader(&buf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			t.Fatal(err)
		}
		_, ok := expected[hdr.Name]
		expected[hdr.Name] = ok
	}

	for k, v := range expected {
		if !v {
			t.Errorf("Got failed file %s", k)
		}
	}
}
