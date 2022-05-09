// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Implements go/functions_framework buildpack.
// The functions_framework buildpack converts a functionn into an application and sets up the execution environment.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
)

const (
	layerName = "quanxiang-plugins"

	quanxiangLowcodeClientPlugins = "plugin-quanxiang-lowcode-client"
	quanxiangLowcodeClientGO      = "plugin-quanxiang-lowcode-client.go"
)

var (
	tmplV0 = template.Must(template.New("plugin-quanxiang-lowcode-client").Parse(quanxiangLowcodeClientTextTemplate))
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(ctx *gcp.Context) (gcp.DetectResult, error) {
	return gcp.OptIn("lowcode plugins"), nil
}

func buildFn(ctx *gcp.Context) error {
	l := ctx.Layer(layerName, gcp.BuildLayer, gcp.CacheLayer)
	ctx.SetFunctionsEnvVars(l)

	ctx.Logf("QUANXIANG lowcode client")

	createPlugins(ctx)
	err := createQuanxiangPlugins(ctx)
	if err != nil {
		return err
	}
	return nil
}

func createPlugins(ctx *gcp.Context) {
	if !ctx.FileExists("plugins") {
		ctx.MkdirAll("plugins", 0755)
	}
}

func createQuanxiangPlugins(ctx *gcp.Context) error {
	fp := filepath.Join("plugins", quanxiangLowcodeClientPlugins)
	os.RemoveAll(fp)
	os.MkdirAll(fp, 0755)

	fd := ctx.CreateFile(filepath.Join(fp, quanxiangLowcodeClientGO))
	defer fd.Close()

	if err := tmplV0.Execute(fd, nil); err != nil {
		return fmt.Errorf("executing template: %v", err)
	}
	return nil
}
