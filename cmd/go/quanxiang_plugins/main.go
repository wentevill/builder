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
	faasLowcode                   = "faas-lowcode"
)

var (
	tmplV0 = template.Must(template.New("plugin-quanxiang-lowcode-client").Parse(quanxiangLowcodeClientTextTemplate))
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(ctx *gcp.Context) (gcp.DetectResult, error) {
	if !ctx.FileExists("go.mod") {
		return gcp.OptOutFileNotFound("go.mod"), nil
	}
	return gcp.OptIn("lowcode plugins"), nil
}

func buildFn(ctx *gcp.Context) error {
	l := ctx.Layer(layerName, gcp.BuildLayer, gcp.CacheLayer)
	ctx.SetFunctionsEnvVars(l)

	// Create quanxiang lowcode client plugins
	createDir(ctx, "plugins")
	err := createQuanxiangPlugins(ctx)
	if err != nil {
		return err
	}

	// Introduce plugin implementation code
	createDir(ctx, "pkg")
	ccp := filepath.Join("pkg", faasLowcode)
	if ctx.FileExists(ccp) {
		ctx.Logf("QUANXIANG lowcode plugin exists")
		return nil
	}

	ctx.Logf("Introduce QUANXIANG lowcode plugin")

	ctx.Exec([]string{"cp", "-R", filepath.Join(ctx.BuildpackRoot(), faasLowcode), ccp})
	ctx.Exec([]string{"echo", ">>", "replace github.com/quanxiang-cloud/faas-lowcode => ./pkg/faas-lowcode"})
	return nil
}

func createDir(ctx *gcp.Context, name string) {
	if !ctx.FileExists(name) {
		ctx.MkdirAll(name, 0755)
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
