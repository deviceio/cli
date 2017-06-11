package hmapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ResourceTestSuite struct {
	suite.Suite
}

func (t *ResourceTestSuite) Test_Resoure_proper_json_marshal_unmarshal() {
	res1 := &Resource{
		Forms: map[string]*Form{
			"AForm": &Form{
				Method:  POST,
				Type:    "WHATEVER",
				Enctype: "WHATEVER",
				Fields: []*FormField{
					&FormField{
						Name:     "WHATEVER",
						Type:     "WHATEVER",
						Encoding: "WHATEVER",
						Required: true,
						Multiple: true,
						Value:    100,
					},
				},
			},
		},
		Links: map[string]*Link{
			"ALink": &Link{},
		},
		Content: map[string]*Content{
			"AContent": &Content{},
		},
	}

	resJSON, err := json.Marshal(res1)

	assert.Nil(t.T(), err)

	var res2 *Resource

	err = json.Unmarshal(resJSON, &res2)

	assert.Nil(t.T(), err)
	assert.NotNil(t.T(), res2)

	//form
	assert.Equal(t.T(), len(res1.Forms), len(res2.Forms))
	aform1, aform2 := res1.Forms["AForm"], res2.Forms["AForm"]
	assert.Equal(t.T(), len(aform1.Fields), len(aform2.Fields))
	assert.Equal(t.T(), aform1.Method, aform2.Method)
	assert.Equal(t.T(), aform1.Type, aform2.Type)
	assert.Equal(t.T(), aform1.Enctype, aform2.Enctype)
	assert.Equal(t.T(), len(aform1.Fields), len(aform2.Fields))
	afield1, afield2 := aform1.Fields[0], aform2.Fields[0]
	assert.Equal(t.T(), afield1.Name, afield2.Name)
	assert.Equal(t.T(), afield1.Type, afield2.Type)
	assert.Equal(t.T(), afield1.Encoding, afield2.Encoding)
	assert.Equal(t.T(), afield1.Required, afield2.Required)
	assert.Equal(t.T(), afield1.Multiple, afield2.Multiple)
	assert.Equal(t.T(), afield1.Value.(int), int(afield2.Value.(float64)))

	//link
	assert.Equal(t.T(), len(res1.Links), len(res2.Links))

	//content
	assert.Equal(t.T(), len(res1.Content), len(res2.Content))
}

func TestResourceTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceTestSuite))
}
