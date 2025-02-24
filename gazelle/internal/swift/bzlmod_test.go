package swift_test

import (
	"testing"

	"github.com/cgrindel/swift_bazel/gazelle/internal/swift"
	"github.com/stretchr/testify/assert"
)

func TestBzlmodStanzas(t *testing.T) {
	awesomeRepoId := "awesome-repo"
	awesomePkg := &swift.Package{
		Name:     swift.RepoNameFromIdentity(awesomeRepoId),
		Identity: awesomeRepoId,
		Remote: &swift.RemotePackage{
			Commit: "12345",
			Remote: "https://github.com/example/awesome-repo",
		},
	}
	anotherRepoID := "another-repo"
	anotherPkg := &swift.Package{
		Name:     swift.RepoNameFromIdentity(anotherRepoID),
		Identity: anotherRepoID,
		Local: &swift.LocalPackage{
			Path: "path/to/another",
		},
	}

	di := swift.NewDependencyIndex()
	di.AddPackage(awesomePkg, anotherPkg)
	di.AddDirectDependency(awesomeRepoId, anotherRepoID)

	actual, err := swift.BzlmodStanzas(di)
	assert.NoError(t, err)
	expected := `swift_deps = use_extension(
    "@cgrindel_swift_bazel//:extensions.bzl",
    "swift_deps",
)
swift_deps.from_file(
    deps_index = "//:swift_deps_index.json",
)
use_repo(
    swift_deps,
    "swiftpkg_another_repo",
    "swiftpkg_awesome_repo",
)
`
	assert.Equal(t, expected, actual)
}
