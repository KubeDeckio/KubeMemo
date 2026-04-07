package assets

import _ "embed"

//go:embed manifests/kubememonote.crd.yaml
var DurableCRDYAML string

//go:embed manifests/kubememoruntimenote.crd.yaml
var RuntimeCRDYAML string

//go:embed manifests/runtime-namespace.yaml
var RuntimeNamespaceYAML string

//go:embed manifests/rbac.yaml
var RBACYAML string
