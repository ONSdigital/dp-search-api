package transformer

import (
	"testing"

	"github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTransform(t *testing.T) {
	t.Parallel()
	expectedItem1 := models.Item{
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
		PopulationType: "PopLbl1",
		Dimensions: []models.ESDimensions{
			{Name: "Dim1", Label: "Lbl1", RawLabel: "RawLbl1"},
		},
	}
	expectedItem2 := models.Item{
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
		PopulationType: "PopLbl2",
		Dimensions: []models.ESDimensions{
			{Name: "Dim2", Label: "Lbl2", RawLabel: "RawLbl2"},
		},
	}
	expectedTopic1 := models.TopicCount{Type: "topic1", Count: 1}

	expectedDimensions := []models.DimensionCount{
		{Type: "dim1", Count: 246},
		{Type: "dim2", Count: 642},
	}

	expectedPopulationTypes := []models.PopulationTypeCount{
		{Type: "pop1", Count: 123},
		{Type: "pop2", Count: 321},
	}

	Convey("Given a new instance of Transformer for ES7x with search responses successfully", t, func() {
		transformer := New()
		esResponse := prepareESMockResponse()

		Convey("When calling a transformer", func() {
			if transformer, ok := transformer.(*Transformer); !ok {
				t.Error("failed to retrieve transfromer")
			} else {
				transformedResponse := transformer.transform(&esResponse, true)
				Convey("Then transforms unmarshalled search responses successfully", func() {
					So(transformedResponse, ShouldNotBeNil)
					So(transformedResponse.Took, ShouldEqual, 10)
					So(len(transformedResponse.Items), ShouldEqual, 2)
					So(transformedResponse.Items[0], ShouldResemble, expectedItem1)
					So(transformedResponse.Items[1], ShouldResemble, expectedItem2)
					So(transformedResponse.Topics[0], ShouldResemble, expectedTopic1)
					So(transformedResponse.Dimensions, ShouldResemble, expectedDimensions)
					So(transformedResponse.PopulationType, ShouldResemble, expectedPopulationTypes)
					So(transformedResponse.Suggestions[0], ShouldResemble, "testSuggestion")
				})
			}
		})
	})
}

func TestTransformDimensions(t *testing.T) {
	Convey("asdf", t, func() {
		counts := models.ESDocCounts{
			Buckets: []models.ESBucket{
				{Key: "dim1", Count: 10},
				{Key: "dim2", Count: 20},
				{Key: "dim3", Count: 30},
			},
		}
		hits := models.ESResponseHits{
			Total: 10,
			Hits: []models.ESResponseHit{
				{
					Source: models.ESSourceDocument{
						Dimensions: []models.ESDimensions{
							{Name: "dim1", Label: "Dimension one"},
							{Name: "dim3", Label: "Dimension three"},
						},
					},
				}, {
					Source: models.ESSourceDocument{
						Dimensions: []models.ESDimensions{
							{Name: "dim2", Label: "Dimension two"},
							{Name: "dim3", Label: "Dimension three"},
						},
					},
				},
			},
		}

		d := transformDimensions(counts, hits)
		So(d, ShouldHaveLength, 3)
		So(d, ShouldContain, models.DimensionCount{Type: "dim1", Label: "Dimension one", Count: 10})
		So(d, ShouldContain, models.DimensionCount{Type: "dim2", Label: "Dimension two", Count: 20})
		So(d, ShouldContain, models.DimensionCount{Type: "dim3", Label: "Dimension three", Count: 30})
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
		PopulationType:  models.ESPopulationType{Name: "Pop1", Label: "PopLbl1"},
		Dimensions: []models.ESDimensions{
			{Name: "Dim1", Label: "Lbl1", RawLabel: "RawLbl1"},
		},
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
		PopulationType:  models.ESPopulationType{Name: "Pop2", Label: "PopLbl2"},
		Dimensions: []models.ESDimensions{
			{Name: "Dim2", Label: "Lbl2", RawLabel: "RawLbl2"},
		},
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

	contentTypeBucket1 := models.ESBucket{
		Key:   "article",
		Count: 1,
	}
	contentTypeBucket2 := models.ESBucket{
		Key:   "product_page",
		Count: 1,
	}

	topicBucket1 := models.ESBucket{
		Key:   "topic1",
		Count: 1,
	}
	topicBucket2 := models.ESBucket{
		Key:   "topic2",
		Count: 1,
	}

	populationTypeBucket1 := models.ESBucket{
		Key:   "pop1",
		Count: 123,
	}

	populationTypeBucket2 := models.ESBucket{
		Key:   "pop2",
		Count: 321,
	}

	dimensionBucket1 := models.ESBucket{
		Key:   "dim1",
		Count: 246,
	}

	dimensionBucket2 := models.ESBucket{
		Key:   "dim2",
		Count: 642,
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
			ContentTypes: models.ESDocCounts{
				Buckets: []models.ESBucket{
					contentTypeBucket1,
					contentTypeBucket2,
				},
			},
			Topic: models.ESDocCounts{
				Buckets: []models.ESBucket{
					topicBucket1,
					topicBucket2,
				},
			},
			PopulationType: models.ESDocCounts{
				Buckets: []models.ESBucket{
					populationTypeBucket1,
					populationTypeBucket2,
				},
			},
			Dimensions: models.ESDocCounts{
				Buckets: []models.ESBucket{
					dimensionBucket1,
					dimensionBucket2,
				},
			},
		},
		Suggest: models.Suggest{
			SearchSuggest: []models.SearchSuggest{
				{Options: []models.Option{{Text: "testSuggestion"}}},
			},
		},
	}

	// Preparing ES response array
	esResponse := models.EsResponses{
		Responses: []*models.EsResponse{
			&esResponse1,
		},
	}

	return esResponse
}
