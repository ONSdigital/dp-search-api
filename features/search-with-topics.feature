Feature: Search endpoint should return data for requested search parameter
    Scenario: When filtering search results with multiple valid topics, I get multiple docs returned. One doc with "canonical_topic" field matching a topic filter and one doc with at least 1 value in "topics" field matching 1 of the topic filters
         Given elasticsearch is healthy
         And elasticsearch returns multiple items in search response with topics filter
         When I GET "/search?topics=123,004"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_multiple_topics_search_result.json"

    Scenario: When filtering search results with multiple valid topics, I get multiple docs returned. Both documents have at least one value in "topics" matching 1 of the topic filters
         Given elasticsearch is healthy
         And elasticsearch returns multiple items in search response with topics filter
         When I GET "/search?topics=002,004"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_multiple_topics_search_result.json"

    Scenario: When filtering search results with multiple valid topics, I get multiple docs returned. Both documents "canonical_topic" matching 1 of the topic filters
         Given elasticsearch is healthy
         And elasticsearch returns multiple items in search response with topics filter
         When I GET "/search?topics=123,456"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_multiple_topics_search_result.json"

    Scenario: When filtering search results with a single valid topic, I only get docs returned that have "canonical_topic" field matching the topic filter
         Given elasticsearch is healthy
         And elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=123"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"

    Scenario: When filtering search results with a single valid topic, I only get docs returned that contain a value in "topics" field matching the topic filter
         Given elasticsearch is healthy
         And elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=001"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"

    Scenario: When filtering search results with multiple topics, 1 invalid and 1 valid topic. I only get docs returned that have "canonical_topic" field matching the valid topic filter
         Given elasticsearch is healthy
         And elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=123,7000"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"

    Scenario: When filtering search results with multiple topics, 1 invalid and 1 valid topic. I only get docs returned that contain a value in "topics" field matching the valid topic filter
         Given elasticsearch is healthy
         And elasticsearch returns one item in search response with topics filter
         When I GET "/search?topics=001,7000"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_single_topic_search_result.json"

    Scenario: When filtering search results with multiple valid topics, I get multiple docs returned and also a distinct count of the topics.
         Given elasticsearch is healthy
         And elasticsearch returns multiple items with distinct topic count in search response
         When I GET "/search?topics=123,004"
         Then the HTTP status code should be "200"
         And the response header "Content-Type" should be "application/json;charset=utf-8"
         And the response body is the same as the json in "./features/testdata/expected_multiple_topics_search_result_with_distinct_topic_count.json"

