Feature: Search by URIs endpoint should return data for requested search parameter

  Scenario: When searching with valid URIs, I get the expected results
    Given elasticsearch is healthy
    And elasticsearch returns one item in search/uris response
    When I POST "/search/uris"
      """
      {
        "uris": ["/peoplepopulationandcommunity"]
      }
      """
    Then the HTTP status code should be "200"
    And the response header "Content-Type" should be "application/json;charset=utf-8"
    And the response body is the same as the json in "./features/testdata/expected_single_search_result.json"

  Scenario: When searching with no URIs provided, I get a bad request response
    Given elasticsearch is healthy
    And elasticsearch returns one item in search/uris response
    When I POST "/search/uris"
      """
      {
        "uris": []
      }
      """
    Then the HTTP status code should be "400"
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And I should receive the following response:
      """
        No URIs provided
      """

  Scenario: When searching with invalid URI format, I get a bad request response
    Given elasticsearch is healthy
    And elasticsearch returns one item in search/uris response
    When I POST "/search/uris"
      """
      {
        "uris": [""]
      }
      """
    Then the HTTP status code should be "400"
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And I should receive the following response:
      """
        Invalid URI: URI cannot be empty
      """

  Scenario: When Elasticsearch returns an error, I get an internal server error
    Given elasticsearch is healthy
    When I POST "/search/uris"
      """
      {
        "uris": ["/peoplepopulationandcommunity"]
      }
      """
    Then the HTTP status code should be "500"
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
    And I should receive the following response:
      """
        call to elastic multisearch api failed
      """
