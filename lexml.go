package lexml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
    "slices"
)

// Set is the entrypoint struct to interact with this program with Fields as a list of Tag types, Content that represent the whole XML file, Raw wich represent the raw byte data of the file
type Set struct {
	Fields  []Tag
	Content Data
	Raw     []byte
}

// Data is a recursive representation of a parsed XML file content
type Data struct {
	Type   *Tag
	Index  int
	Value  string
	Inners []Data
}

// DataAlt is an altered Data Type without the Inners element, it's used when returning a list of specific Data elements from a LookupId, LookupName or LookupIndex functions
type DataAlt struct {
    Type *Tag
    Index int
    Value string
}

// Tags represent fields of a XML file, those are generated on the go
type Tag struct {
	Id   int
	Name string
}

// Parse convert the whole file into a DataSet datastructure
func (set Set) Parse() ([]Tag, Data, error) {

	reader := bytes.NewReader(set.Raw)
	decoder := xml.NewDecoder(reader)

	tagList := findTags(decoder)

	reader = bytes.NewReader(set.Raw)
	decoder = xml.NewDecoder(reader)

	content, err := genData(decoder, tagList)
	if err != nil {
		return []Tag{}, Data{}, fmt.Errorf("Parse() -> %s", err)
	}

	return tagList, content, nil
}

func (data Data) Alter() DataAlt {
    return DataAlt{
        Type: data.Type,
        Index: data.Index,
        Value: data.Value,
    }
}

/*

This won't work since Fields are not accessible, i should consider modify those Data methods to be using Set instead, into each recursion, i should make a copy of the given Set with the Content modified to be the inner content as if we are into recursion.

*/

// LookupId search recursively throught the Content of a given Set Type and returns a list of pointers to every Data elements that are equal to the given id
func (data Data) LookupId(id int) []DataAlt {

    dataList := make([]DataAlt, 0)
    givenId := 0

    if intEq(id, data.Type.Id) {
        dataList = append(dataList, data.Alter())
    }

    for index, data := range data.Inners {

        givenId = data.Type.Id

        if intEq(id, givenId) {
            dataList = append(dataList, data.Inners[index].Alter())
        }

        dataList = slices.Concat(dataList, data.Inners[index].LookupId(index))
    }

    return dataList
}

// LookupName search recursively throught the Content of a given Set Type and returns a list of pointers to every Data elements that are equal to the given id
func (data Data) LookupName(name string) []DataAlt {

    dataList := make([]DataAlt, 0)
    givenName := ""

    if strEq(name, data.Type.Name) {
        dataList = append(dataList, data.Alter())
    }

    for index, data := range data.Inners {
        
        givenName = data.Type.Name

        if strEq(name, givenName) {
            dataList = append(dataList, data.Inners[index].Alter())
        }

        dataList = slices.Concat(dataList, data.Inners[index].LookupId(index))
    }

    return[]DataAlt{}
}

// LookupIndex search recursively througt the Content of a given Set Type and returns the element present at a specific index of a given depth
func (set Set) LookupIndex(depth int, index int) []DataAlt {
    return []DataAlt{}
}

// NewSet generates a set that's retuned as pointer
func NewSet(buff []byte) *Set {

	return &Set{

		Fields: make([]Tag, 0),

		Content: Data{
			Type:   nil,
			Index:  0,
			Value:  "",
			Inners: make([]Data, 0),
		},

		Raw: buff,
	}
}

// recurse recreates a recursive Data datastructure representation of the file itself. It's taking every tags, subtags and data to make them one recursive datastructure of Type Data
func genData(decoder *xml.Decoder, tagList []Tag) (Data, error) {

	data := newData(0)
	index := 0

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}

		switch tk := tok.(type) {

		case xml.StartElement:
			name := tk.Name.Local

			data.Inners = append(data.Inners, newData(index))

			data.Inners[index], err = genData(decoder, tagList)
			if err != nil {
				return Data{}, fmt.Errorf("recurse() -> %s", err)
			}

			data.Inners[index].Type = getTag(tagList, name)

			index = index + 1

		case xml.EndElement:
			return data, nil

		case xml.CharData:
			data.Value = string(tk)

		default:
			return Data{}, fmt.Errorf("recurse() -> Unknown or unused Type encountered, got %T", tok)
		}
	}

	return data, nil
}

// newData return an empty Data Type
func newData(index int) Data {

	return Data{
        Type:   &Tag{Id: 0, Name: "XMLROOT"},
		Index:  index,
		Value:  "",
		Inners: make([]Data, 0),
	}
}

// findTags find and return the exhaustive list of unique tags found in the xml file
func findTags(decoder *xml.Decoder) []Tag {

	tagList := make([]Tag, 0)
	index := 0

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}

		switch tk := tok.(type) {

		case xml.StartElement:

			if !tagExist(tk.Name.Local, tagList) {

				tagList = append(tagList, Tag{
					Id:   index,
					Name: tk.Name.Local,
				})
			}
		}
	}
	return tagList
}

// getTag return a single tag from a tagList with a given name as string
func getTag(tagList []Tag, tok string) *Tag {

	for index, tag := range tagList {
		if tok == tag.Name {
			return &tagList[index]
		}
	}

	return nil
}

// returns the Id of a given Data Type
func getId(data Data) int {
    return data.Type.Id
}

// intEq checks equality of two given Int values 
func intEq(orig int, given int) bool {
    return (orig == given)
}

// returns the Name of a given Data Type
func getName(data Data) int {
    return data.Type.Id
}

// strEq checks equality of two given String values
func strEq(orig string, given string) bool {
    return bytes.Equal([]byte(orig), []byte(given))
}

// tagExist checks if a tag exist in the whole list of tags
func tagExist(word string, tagList []Tag) bool {

	for _, tag := range tagList {

		if tag.Name == word {
			return true
		}
	}

	return false
}
