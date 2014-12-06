package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
}

func NewToRocket(name, version, os, arch string) (*ToRocket, error) {
	t := &ToRocket{}
	if err := t.setBasicData(name, version, os, arch); err != nil {
		return nil, err
	}

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
