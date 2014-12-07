package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/rocket/app-container/schema"
	"github.com/coreos/rocket/app-container/schema/types"
	"github.com/docker/docker/builder/parser"
)

const (
	DEFAULT_AC_KIND    = "AppManifest"
	DEFAULT_AC_VERSION = "1.0.0"
)

type ToRocket struct {
	manifest schema.AppManifest
	aci      *ACIFile
}

func NewToRocket(name, version, os, arch string) (*ToRocket, error) {
	t := &ToRocket{}
	if err := t.setBasicData(name, version, os, arch); err != nil {
		return nil, err
	}

	t.aci = NewACIFile()

	return t, nil
}

func (t *ToRocket) setBasicData(name, version, os, arch string) error {
	t.manifest.Name = types.ACName(name)
	t.manifest.Version = types.ACName(version)
	t.manifest.OS = types.ACName(os)
	t.manifest.Arch = types.ACName(arch)

	t.manifest.ACKind = types.ACKind(DEFAULT_AC_KIND)

	if ver, err := types.NewSemVer(DEFAULT_AC_VERSION); err == nil {
		t.manifest.ACVersion = *ver
	} else {
		return err
	}

	return nil
}

func (t *ToRocket) Process(n *parser.Node) {
	t.processNode(n)
}

func (t *ToRocket) processNode(n *parser.Node) {
	switch n.Value {
	case "add":
		t.processAddNode(n)
	case "cmd":
		t.processCMDNode(n)
	case "volume":
		t.processVolumeNode(n)
	case "env":
		t.processEnvNode(n)
	case "expose":
		t.processExposeNode(n)
	}

	if len(n.Children) != 0 {
		t.iterateNodes(n.Children)
	}
}

func (t *ToRocket) processAddNode(n *parser.Node) {
	add := n.Original[4:]
	files := strings.Split(add, " ")
	dst := files[len(files)-1]
	if dst[len(dst)-1] != '/' {
		dst += "/"
	}

	dst = "rootfs/" + dst
	for _, file := range files[:len(files)-1] {
		t.aci.AddFile(file, dst)
	}
}

func (t *ToRocket) processCMDNode(n *parser.Node) {
	cmd := n.Original[4:]
	if isJSON, ok := n.Attributes["json"]; ok && isJSON {
		var data []string
		json.Unmarshal([]byte(n.Original[4:]), &data)
		cmd = strings.Join(data, " ")
	}

	t.manifest.Exec = []string{cmd}
}

func (t *ToRocket) processVolumeNode(n *parser.Node) {
	var volumes []string

	if isJSON, ok := n.Attributes["json"]; ok && isJSON {
		json.Unmarshal([]byte(n.Original[7:]), &volumes)
	} else {
		volumes = []string{n.Original[7:]}
	}

	t.manifest.MountPoints = make([]types.MountPoint, len(volumes))
	for i, path := range volumes {
		pathS := strings.Split(path, "/")
		t.manifest.MountPoints[i] = types.MountPoint{
			Name: types.ACName(pathS[len(pathS)-1]),
			Path: path,
		}
	}
}

func (t *ToRocket) processEnvNode(n *parser.Node) {
	env := n.Original[4:]
	values := strings.Split(env, " ")

	if t.manifest.Environment == nil {
		t.manifest.Environment = make(map[string]string)
	}

	t.manifest.Environment[values[0]] = values[1]
}

func (t *ToRocket) processExposeNode(n *parser.Node) {
	expose := n.Original[7:]
	port := strings.Split(expose, "/")
	portInt, _ := strconv.Atoi(port[0])

	proto := "tcp"
	if len(port) == 2 {
		proto = port[1]
	}

	t.manifest.Ports = append(t.manifest.Ports, types.Port{
		Name:     types.ACName(port[0]),
		Protocol: proto,
		Port:     uint(portInt),
	})
}

func (t *ToRocket) iterateNodes(nodes []*parser.Node) {
	for _, n := range nodes {
		t.Process(n)
	}
}

func (t *ToRocket) Print() {
	json, _ := t.manifest.MarshalJSON()
	fmt.Printf("%s", json)
}

func (t *ToRocket) SaveToFile(filename string) error {
	json, err := t.manifest.MarshalJSON()
	if err != nil {
		return err
	}

	if err := t.aci.AddFileFromBytes("app", json); err != nil {
		return err
	}

	if err := t.aci.SaveToFile(filename); err != nil {
		return err
	}

	return nil
}

type ACIFile struct {
	contents []*content
}

type content struct {
	header *tar.Header
	raw    []byte
}

func NewACIFile() *ACIFile {
	return &ACIFile{make([]*content, 0)}
}

func (a *ACIFile) AddFileFromBytes(filename string, raw []byte) error {
	c := &content{}

	time := time.Now()

	c.header = &tar.Header{
		Name:       filename,
		Size:       int64(len(raw)),
		ModTime:    time,
		AccessTime: time,
		ChangeTime: time,
	}

	c.raw = raw
	a.contents = append(a.contents, c)

	return nil
}

func (a *ACIFile) AddFile(src string, dst string) error {
	c := &content{}

	var err error
	c.raw, err = ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	fInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	c.header, err = tar.FileInfoHeader(fInfo, "")
	c.header.Name = dst + path.Base(src)
	if err != nil {
		return err
	}

	a.contents = append(a.contents, c)

	return nil
}

func (a *ACIFile) SaveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	g := gzip.NewWriter(f)
	t := tar.NewWriter(g)

	defer func() {
		t.Close()
		g.Close()
		f.Close()
	}()

	for _, c := range a.contents {
		if err := t.WriteHeader(c.header); err != nil {
			return err
		}

		if _, err := t.Write(c.raw); err != nil {
			return err
		}
	}

	return nil
}
