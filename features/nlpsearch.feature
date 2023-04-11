Feature: NLP
Scenario: When Searching for Without OAC or SIC codes I get empty resp
    When I GET "/nlp/search?q=dentists"
    Then the HTTP status code should be "200"
    And the response header "Content-Type" should be "application/json;charset=utf-8"
    And nlp response is the same as in "./features/testdata/nlpdata/emptyResponse.json"