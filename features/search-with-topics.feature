Feature: Search endpoint should return data for requested search parameter
    Scenario: When Searching for a canonical topic and a sub topic topics I get multiple docs
         Given elasticsearch returns multiple items in search response with topics filter
         When I GET "/search?topics=123,004"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_multiple_topics_search_result.json"

    Scenario: When Searching for multiple sub topics I get multiple docs
         Given elasticsearch returns multiple items in search response with topics filter
         When I GET "/search?topics=002,004"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_multiple_topics_search_result.json"

    Scenario: When Searching for multiple canonical topics I get multiple docs
         Given elasticsearch returns multiple items in search response with topics filter
         When I GET "/search?topics=123,456"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_multiple_topics_search_result.json"

    Scenario: When Searching for single canonical topic I get single doc
         Given elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=123"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"

    Scenario: When Searching for single sub topic I get single doc
         Given elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=001"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"

    Scenario: When Searching for multiple canonical topics, one that exists and another that does not exist I get single doc
         Given elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=123,7000"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"

    Scenario: When Searching for multiple sub topics, one that exists and another that does not exist I get single doc
         Given elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=001,7000"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"



