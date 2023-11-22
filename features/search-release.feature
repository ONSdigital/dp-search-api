Feature: search/releases endpoint should return data for various combinations of input parameters
  Scenario: When Searching for published releases relating to 'Education in Wales' between certain dates I get one result
    Given elasticsearch is healthy
    And elasticsearch returns one item in search/release response
    When I GET "/search/releases?q=Education+in+Wales&fromDate=2020-01-01&toDate=2020-12-31&release-type=type-published"
    And the HTTP status code should be "200"
    And the response header "Content-Type" should be "application/json;charset=utf-8"
    And the response body is the same as the json in "./features/testdata/expected_single_search_release_result.json"

  Scenario: When Searching for published releases relating to 'Education in Wales' between certain dates with highlighting turned off and I get one result
    Given elasticsearch is healthy
    And elasticsearch returns one item in search/release response
    When I GET "/search/releases?q=Education+in+Wales&response=2020-01-01&toDate=2020-12-31&release-type=type-published&highlight=false"
    Then the HTTP status code should be "200"
    And the response header "Content-Type" should be "application/json;charset=utf-8"
    And the response body is the same as the json in "./features/testdata/expected_single_search_release_result_nohighlight.json"

  Scenario: When Searching for published releases relating to 'Education in Wales' between certain dates I get multiple results
    Given elasticsearch is healthy
    And elasticsearch returns multiple items in search/release response
    When I GET "/search/releases?q=Education+in+Wales&response=2020-01-01&toDate=2020-12-31&release-type=type-published"
    Then the HTTP status code should be "200"
    And the response header "Content-Type" should be "application/json;charset=utf-8"
    And the response body is the same as the json in "./features/testdata/expected_multiple_search_release_results.json"

  Scenario: When Searching for published releases relating to 'Education in Wales' between certain dates I get zero results
    Given elasticsearch is healthy
    And elasticsearch returns zero items in search/release response
    When I GET "/search/releases?q=Education+in+Scotland&response=2020-01-01&toDate=2020-12-31&release-type=type-published"
    Then the HTTP status code should be "200"
    And the response header "Content-Type" should be "application/json;charset=utf-8"
    And the response body is the same as the json in "./features/testdata/expected_zero_search_release_results.json"

  Scenario: When Searching for published releases relating to 'Education in Wales' between certain dates, I get internal server error
    Given elasticsearch is healthy
    And elasticsearch returns internal server error
    When I GET "/search/releases?q=Education+in+Scotland&response=2020-01-01&toDate=2020-12-31&release-type=type-published"
    Then the HTTP status code should be "500"
    And the response header "Content-Type" should be "text/plain; charset=utf-8"
