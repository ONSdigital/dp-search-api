Feature: Search endpoint should return data for requested search parameter
     Scenario: When filtering search results with a valid dataset ID, I get one result returned that is associated with that dataset ID.
          Given elasticsearch is healthy
          And elasticsearch returns one item in search response with dataset ID
          When I GET "/search?dataset_id=TS056"
          Then the HTTP status code should be "200"
          And the response header "Content-Type" should be "application/json;charset=utf-8"
          And the response body is the same as the json in "./features/testdata/expected_single_dataset_search_result.json"


     Scenario: When filtering search results with an invalid dataset ID, I get no documents returned.
          Given elasticsearch is healthy
          When I GET "/search?dataset_id=q"
          Then the HTTP status code should be "400"
          And the response header "Content-Type" should be "text/plain; charset=utf-8"
          And I should receive the following response:
            """
            Invalid DatasetID(s): q
            """


     Scenario: When filtering search results with multiple dataset IDs (one valid dataset ID and one invalid dataset ID), and there are no results for the valid dataset ID, the response should indicate no results.
          Given elasticsearch is healthy
          When I GET "/search?dataset_id=TS056,d"
          Then the HTTP status code should be "400"
          And the response header "Content-Type" should be "text/plain; charset=utf-8"
          And I should receive the following response:
            """
            Invalid DatasetID(s): d
            """


     Scenario: When filtering search results with valid dataset ID but no matches I get zero results
          Given elasticsearch is healthy
          And elasticsearch returns zero items in search response
          When I GET "/search?q=RPI%20Consumers&dataset_id=QNA"
          Then the HTTP status code should be "200"
          And the response header "Content-Type" should be "application/json;charset=utf-8"
          And the response body is the same as the json in "./features/testdata/expected_zero_search_results.json"


     Scenario: When filtering search results with multiple dataset IDs, I get multiple results that are associated with any of the provided dataset IDs.
          Given elasticsearch is healthy
          And elasticsearch returns multiple items in search response for the dataset IDs
          When I GET "/search?dataset_id=dataset123,TS056"
          Then the HTTP status code should be "200"
          And the response header "Content-Type" should be "application/json;charset=utf-8"
          And the response body is the same as the json in "./features/testdata/expected_multiple_dataset_search_result.json"
