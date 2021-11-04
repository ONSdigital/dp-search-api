Feature: Search endpoint should return data for requested search parameter
    Scenario: When Searching for CPI I get the same results as Consumer Price Index
        Given elasticsearch returns multiple items in search response
        When I GET "/search?q=consumer%20price%20index"
        Then the HTTP status code should be "200"
        And the response header "Content-Type" should be "application/json;charset=utf-8"
        And the response body is the same as the json in "./features/testdata/multiple_search_expected_highlighted.json"
        