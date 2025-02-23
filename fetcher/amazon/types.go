package amazon

import "encoding/xml"

// root is a struct of releasemd.xml for AL2022
// curl https://al2022-repos-us-west-2-9761ab97.s3.dualstack.us-west-2.amazonaws.com/core/releasemd.xml
type root struct {
	XMLName  xml.Name `xml:"root"`
	Releases struct {
		Release []struct {
			Version string `xml:"version,attr"`
			Update  []struct {
				Name          string `xml:"name"`
				VersionString string `xml:"version_string"`
				ReleaseNotes  string `xml:"release_notes"`
			} `xml:"update"`
		} `xml:"release"`
	} `xml:"releases"`
}

// extrasCatalog is a struct of extras-catalog.json for Amazon Linux 2 Extra Repository
type extrasCatalog struct {
	Topics []struct {
		N            string   `json:"n"`
		Inst         []string `json:"inst,omitempty"`
		Versions     []string `json:"versions"`
		DeprecatedAt string   `json:"deprecated-at,omitempty"`
		Visible      []string `json:"visible,omitempty"`
	} `json:"topics"`
}

// repoMd has repomd data
type repoMd struct {
	RepoList []repo `xml:"data"`
}

// repo has a repo data
type repo struct {
	Type     string   `xml:"type,attr"`
	Location location `xml:"location"`
}

// location has a location of repomd
type location struct {
	Href string `xml:"href,attr"`
}
