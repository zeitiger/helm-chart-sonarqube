package tests

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// k8s label values must match this and be <= 63 chars.
var labelValueRegexp = regexp.MustCompile(`^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$`)

const digest = "@sha256:0b3408443fb6a311253de27941b10977c9b977fd53be2265947bfe14161308b8"

// A pinned digest in image.tag (tag@sha256:...) must not leak into the
// app.kubernetes.io/version label, which would make the StatefulSet invalid,
// while the container image reference must keep the full digest.
func TestVersionLabelStripsDigest(t *testing.T) {
	opts := newSQHelmOptions()
	opts.SetValues["image.tag"] = "2026.5.0-developer" + digest

	rendered := renderSQStsTemplate(t, "test-cases-values/sonarqube/version-label.yaml", opts)

	version := rendered.Labels["app.kubernetes.io/version"]
	assert.Equal(t, "2026.5.0-developer", version)
	assert.LessOrEqual(t, len(version), 63)
	assert.Regexp(t, labelValueRegexp, version)

	// The image reference itself keeps the digest, so pinning is retained.
	image := rendered.Spec.Template.Spec.Containers[0].Image
	assert.True(t, strings.HasSuffix(image, digest),
		"container image should keep the pinned digest, got %q", image)
}

// A plain tag without a digest is passed through unchanged.
func TestVersionLabelPlainTag(t *testing.T) {
	opts := newSQHelmOptions()
	opts.SetValues["image.tag"] = "2026.5.0-developer"

	rendered := renderSQStsTemplate(t, "test-cases-values/sonarqube/version-label.yaml", opts)

	version := rendered.Labels["app.kubernetes.io/version"]
	require.Equal(t, "2026.5.0-developer", version)
	assert.Regexp(t, labelValueRegexp, version)
}
