package artifacts

import "encoding/xml"

// ArtifactPosition holds the position of the artifact in the XML file
// hierarchy ex: API -> Resource -> InSequence -> Log
type ArtifactPosition struct {
	LineNo    int
	FileName  string
	Hierarchy string
}

type SubArtifact interface {
	UnmarshalArtifact(d *xml.Decoder, start xml.StartElement, position ArtifactPosition) error
}

type RootArtifact interface {
	UnmarshalArtifact(fileContent string, position ArtifactPosition) error
}
