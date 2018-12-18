package microformats_test

import (
	"fmt"
	"net/url"
	"strings"

	microformats "hawx.me/code/microformats"
)

func ExampleParse() {
	r := strings.NewReader(`<div class="h-card">
  <span class="p-name">John Doe</span>
</div>`)
	baseURL, _ := url.Parse("http://example.com/")

	var output struct {
		Name string `mf:"name"`
	}

	microformats.Parse(r, baseURL, "h-card", &output)
	fmt.Println(output.Name)
	// Output: John Doe
}

func ExampleParseAll() {
	r := strings.NewReader(`<div class="h-card">
  <span class="p-name">John Doe</span>
</div>
<div class="h-card">
  <span class="p-name">Jane Doe</span>
</div>`)
	baseURL, _ := url.Parse("http://example.com/")

	var output []struct {
		Name string `mf:"name"`
	}

	microformats.ParseAll(r, baseURL, "h-card", &output)
	fmt.Println(output[0].Name)
	fmt.Println(output[1].Name)
	// Output: John Doe
	// Jane Doe
}
