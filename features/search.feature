Feature: Search endpoint should return data for requested search parameter
    Scenario: When Searching for CPIH I get one result
        Given elasticsearch is healthy
        And elasticsearch returns one item in search response
        When I GET "/search?q=CPIH"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/expected_single_search_result.json"

    Scenario: When Searching for CPI I get multiple results
        Given elasticsearch is healthy
        And elasticsearch returns multiple items in search response
        When I GET "/search?q=CPI"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/expected_multiple_search_results.json"

    Scenario: When Searching for RPI I get zero results
        Given elasticsearch is healthy
        And elasticsearch returns zero items in search response
        When I GET "/search?q=RPI%20Consumers"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/expected_zero_search_results.json"

    Scenario: When Searching for CPI, I get internal server error
        Given elasticsearch is healthy
        And elasticsearch returns internal server error
        When I GET "/search?q=CPI"
        Then the HTTP status code should be "500"
        And the response header "Content-Type" should be "text/plain; charset=utf-8"

    Scenario: When Searching with invalid characters I get a bad request response
        Given elasticsearch is healthy
        When I GET "/search?q=tiktok%E6%80%8E%E4%B9%88%E5%BC%80%E9"
        Then the HTTP status code should be "400"
        And the response header "Content-Type" should be "text/plain; charset=utf-8"
        And I should receive the following response:
            """
            invalid characters in query
            """

    Scenario: When Searching with a valid uri_prefix I get multiple results
        Given elasticsearch is healthy
        And elasticsearch returns multiple items in search response
        When I GET "/search?uri_prefix=/releases"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/expected_multiple_search_results.json"

    Scenario: When Searching with a invalid uri_refix I get a bad request response
        Given elasticsearch is healthy
        When I GET "/search?uri_prefix=economy"
        Then the HTTP status code should be "400"
        And the response header "Content-Type" should be "text/plain; charset=utf-8"
        And I should receive the following response:
            """
            invalid URI prefix parameter
            """
