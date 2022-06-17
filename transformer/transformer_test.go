package transformer

import (
	"testing"

	"github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTransform(t *testing.T) {
	t.Parallel()
	expectedESDocument1 := models.ESSourceDocument{
		DataType:        "anyDataType1",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword1"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic1"},
		Highlight: &models.HighlightObj{
			DatasetID: "",
		},
	}
	expectedESDocument2 := models.ESSourceDocument{
		DataType:        "anyDataType2",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword2"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic2"},
		Highlight: &models.HighlightObj{
			DatasetID: "",
		},
	}

	Convey("Given a new instance of Transformer for ES7x with search responses successfully", t, func() {
		transformer := New()
		esResponse := prepareESMockResponse()

		Convey("When calling a transformer", func() {
			if transformer, ok := transformer.(*Transformer); !ok {
				t.Error("failed to retrieve legacy transfromer")
			} else {
				transformedResponse := transformer.transform(&esResponse, true)
				Convey("Then transforms unmarshalled search responses successfully", func() {
					So(transformedResponse, ShouldNotBeNil)
					So(transformedResponse.Took, ShouldEqual, 10)
					So(len(transformedResponse.Items), ShouldEqual, 2)
					So(transformedResponse.Items[0], ShouldResemble, expectedESDocument1)
					So(transformedResponse.Items[1], ShouldResemble, expectedESDocument2)
					So(transformedResponse.Suggestions[0], ShouldResemble, "testSuggestion")
				})
			}
		})
	})
}

// Prepare mock ES response
func prepareESMockResponse() models.EsResponses {
	esDocument1 := models.ESSourceDocument{
		DataType:        "anyDataType1",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword1"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic1"},
	}

	esDocument2 := models.ESSourceDocument{
		DataType:        "anyDataType2",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword2"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic2"},
	}

	esDocuments := []models.ESSourceDocument{esDocument1, esDocument2}

	hit := models.ESResponseHit{
		Source:    esDocuments[0],
		Highlight: &models.ESHighlight{},
	}

	hit2 := models.ESResponseHit{
		Source:    esDocuments[1],
		Highlight: &models.ESHighlight{},
	}

	bucket1 := models.ESBucket{
		Key:   "article",
		Count: 1,
	}
	bucket2 := models.ESBucket{
		Key:   "product_page",
		Count: 1,
	}
	buckets := []models.ESBucket{bucket1, bucket2}

	esDoccount := models.ESDocCounts{
		Buckets: buckets,
	}

	esResponse1 := models.EsResponse{
		Took: 10,
		Hits: models.ESResponseHits{
			Hits: []models.ESResponseHit{
				hit,
				hit2,
			},
		},
		Aggregations: models.ESResponseAggregations{
			DocCounts: esDoccount,
		},
		Suggest: models.Suggest{
			SearchSuggest: []models.SearchSuggest{
				{Text: "testSuggestion"},
			},
		},
	}

	// Preparing ES response array
	esResponse := models.EsResponses{
		Responses: []models.EsResponse{
			esResponse1,
		},
	}

	return esResponse
}
