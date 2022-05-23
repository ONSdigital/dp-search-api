Feature: Search endpoint should return data for requested search parameter
    Scenario: When Searching for CPIH I get one result
        Given elasticsearch7x returns one item in search response
        When I GET "/search?q=CPIH"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./es710_features/testdata/expected_single_search_result.json"

    Scenario: When Searching for CPI I get multiple results
        Given elasticsearch7x returns multiple items in search response
        When I GET "/search?q=CPI"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./es710_features/testdata/expected_multiple_search_results.json"

    Scenario: When Searching for RPI I get zero results
        Given elasticsearch7x returns zero items in search response
        When I GET "/search?q=RPI%20Consumers"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./es710_features/testdata/expected_zero_search_results.json"

    Scenario: When Searching for CPI, I get internal server error
        Given elasticsearch7x returns internal server error
        When I GET "/search?q=CPI"
        Then the HTTP status code should be "500"
        And the response header "Content-Type" should be "text/plain; charset=utf-8"