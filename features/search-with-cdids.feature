Feature: Search endpoint should return data for requested cdid parameter

    Scenario: When Searching with valid cdid I get one result
        Given elasticsearch is healthy
        And elasticsearch returns one item in search response
        When I GET "/search?cdids=ABC1"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/expected_single_search_result.json"

    Scenario: When Searching with multiple cdid values I get multiple results
        Given elasticsearch is healthy
        And elasticsearch returns multiple items in search response
        When I GET "/search?cdids=ABC12,ABC13"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/expected_multiple_search_results.json"

    Scenario: When Searching with invalid cdid I get a bad request response
        Given elasticsearch is healthy
        When I GET "/search?cdids=INVALID"
        Then the HTTP status code should be "400"
        And the response header "Content-Type" should be "text/plain; charset=utf-8"
        And I should receive the following response:
            """
            Invalid cdid(s): INVALID
            """
    Scenario: When Searching with multiple invalid cdids I get a bad request response
        Given elasticsearch is healthy
        When I GET "/search?cdids=INVALID1,INVALID2"
        Then the HTTP status code should be "400"
        And the response header "Content-Type" should be "text/plain; charset=utf-8"
        And I should receive the following response:
            """
            Invalid cdid(s): INVALID1,INVALID2
            """

    Scenario: When Searching with valid cdid but no matches I get zero results
        Given elasticsearch is healthy
        And elasticsearch returns zero items in search response
        When I GET "/search?cdids=BAD1"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/expected_zero_search_cdid_results.json"