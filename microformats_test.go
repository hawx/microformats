package microformats

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var baseURL, _ = url.Parse("http://example.com/")

func TestParseWithUntaggedField(t *testing.T) {
	assert := assert.New(t)
	input := strings.NewReader(`<div class="h-card"><span class="p-name">John Doe</span></div>`)
	var output struct {
		Name string
	}

	err := Parse(input, baseURL, "h-card", &output)
	assert.Nil(err)
	assert.Equal("", output.Name)
}

func TestParseWithStringField(t *testing.T) {
	assert := assert.New(t)
	input := strings.NewReader(`<div class="h-card"><span class="p-name">John Doe</span></div>`)
	var output struct {
		Name string `mf:"name"`
	}

	err := Parse(input, baseURL, "h-card", &output)
	assert.Nil(err)
	assert.Equal("John Doe", output.Name)
}

func TestParseWithArrayField(t *testing.T) {
	assert := assert.New(t)
	input := strings.NewReader(`<div class="h-card">
<span class="p-name">John Doe</span>
but sometimes <span class="p-name">Jonathan Doe</span>
</div>`)
	var output struct {
		Name []string `mf:"name"`
	}

	err := Parse(input, baseURL, "h-card", &output)
	assert.Nil(err)
	if assert.Len(output.Name, 2) {
		assert.Equal("John Doe", output.Name[0])
		assert.Equal("Jonathan Doe", output.Name[1])
	}
}

func TestParseWithMissingField(t *testing.T) {
	assert := assert.New(t)
	input := strings.NewReader(`<div class="h-card"><span class="p-name">John Doe</span></div>`)
	var output struct{}

	err := Parse(input, baseURL, "h-card", &output)
	assert.Nil(err)
}

func TestParseWithUnknownField(t *testing.T) {
	assert := assert.New(t)
	input := strings.NewReader(`<div class="h-card"><span class="p-name">John Doe</span></div>`)
	var output struct {
		Name string `mf:"name"`
		What string `mf:"what"`
	}

	err := Parse(input, baseURL, "h-card", &output)
	assert.Nil(err)
	assert.Equal("John Doe", output.Name)
	assert.Equal("", output.What)
}

type MyHCard struct {
	Name   string   `mf:"name"`
	Photos []string `mf:"photo"`
}

func TestParseCustomStructure(t *testing.T) {
	assert := assert.New(t)

	input := strings.NewReader(`<div class="h-card">
  <a class="u-url p-name" href="/john">John Doe</a>
  <a class="u-email" href="mailto:john@example.com">john@example.com</a>
  <img class="u-photo" src="/photo.jpg" />
</div>`)

	var output MyHCard
	err := Parse(input, baseURL, "h-card", &output)
	assert.Nil(err)
	assert.Equal("John Doe", output.Name)
	if assert.Len(output.Photos, 1) {
		assert.Equal("http://example.com/photo.jpg", output.Photos[0])
	}
}

func TestParseAllCustomStructures(t *testing.T) {
	assert := assert.New(t)

	input := strings.NewReader(`<div class="h-card">
  <a class="u-url p-name" href="/john">John Doe</a>
  <a class="u-email" href="mailto:john@example.com">john@example.com</a>
  <img class="u-photo" src="/photo.jpg" />
</div>
<div class="h-card">
  <a class="u-url p-name" href="/jane">Jane Doe</a>
  <a class="u-email" href="mailto:jane@example.com">jane@example.com</a>
  <img class="u-photo" src="http://other.example.com/photo2.jpg" />
</div>`)

	var output []MyHCard
	err := ParseAll(input, baseURL, "h-card", &output)
	assert.Nil(err)

	if assert.Len(output, 2) {
		assert.Equal("John Doe", output[0].Name)
		if assert.Len(output[0].Photos, 1) {
			assert.Equal("http://example.com/photo.jpg", output[0].Photos[0])
		}

		assert.Equal("Jane Doe", output[1].Name)
		if assert.Len(output[1].Photos, 1) {
			assert.Equal("http://other.example.com/photo2.jpg", output[1].Photos[0])
		}
	}
}

func TestParseWithNestedStructure(t *testing.T) {
	assert := assert.New(t)

	input := strings.NewReader(`<div class="h-card">
  <a class="p-name u-url"
     href="http://blog.lizardwrangler.com/"
    >Mitchell Baker</a>
  (<a class="p-org h-card h-org"
      href="http://mozilla.org/"
     >Mozilla Foundation</a>)
</div>`)

	var outputOrg struct {
		Name string `mf:"name"`
		URL  string `mf:"url"`
		Org  struct {
			Name string `mf:"name"`
			URL  string `mf:"url"`
		} `mf:"org"`
	}

	err := Parse(input, baseURL, "h-card", &outputOrg)
	assert.Nil(err)

	assert.Equal("Mitchell Baker", outputOrg.Name)
	assert.Equal("http://blog.lizardwrangler.com/", outputOrg.URL)
	assert.Equal("Mozilla Foundation", outputOrg.Org.Name)
	assert.Equal("http://mozilla.org/", outputOrg.Org.URL)
}
