Feature: Search endpoint should return data for requested search parameter
    Scenario: When Searching for CPI I get the same results as Consumer Price Index
        Given I GET "/search?q=consumer%20price%20index"
        And I should receive the following JSON response with status "200":
        """
            {
                "count":"",
                "took":"530",
                "content_types":[
                    {
                        "type":"",
                        "count":"",
                    }
                ],
                "items": [
                    "description": "",
                    "type": "",
                    "uri": "",
                    ],
                 "suggestions":[""],
                 "alternative_suggestions": [""]
            }
        """
        When I GET "/search?q=cpi"
        Then the response header "Content-Type" should be "application/json; charset=utf-8"
        And I should receive the following JSON response with status "200":
        """
            {
                "count":"",
                "took":"530",
                "content_types":[
                    {
                        "type":"",
                        "count":"",
                    }
                ],
                "items": [
                    "description": "",
                    "type": "",
                    "uri": "",
                    ],
                 "suggestions":[""],
                 "alternative_suggestions": [""]
            }
        """

#    Scenario: Returning a WARNING (429) status when one downstream service is warning
#        Given one of the downstream services is warning
#        And I wait "2" seconds for the healthcheck to be available
#        When I GET "/health"
#        Then the HTTP status code should be "429"
#        And the response header "Content-Type" should be "application/json; charset=utf-8"
#        And I should receive the following health JSON response:
#        """
#            {
#                "status": "WARNING",
#                "version": {
#                    "git_commit": "3t7e5s1t4272646ef477f8ed755",
#                    "language": "go",
#                    "language_version": "go1.16.5",
#                    "version": "v1.2.3"
#                },
#                "checks": [
#                    {
#                        "name": "elasticsearch",
#                        "status": "WARNING",
#                        "status_code": 429,
#                        "message": "elasticsearch is degraded, but at least partially functioning"
#                    }
#                ]
#            }
#        """
#
#    Scenario: Returning a WARNING (429) status when one downstream service is critical and critical timeout has not expired
#        Given one of the downstream services is failing
#        And I wait "2" seconds for the healthcheck to be available
#        When I GET "/health"
#        Then the HTTP status code should be "429"
#        And the response header "Content-Type" should be "application/json; charset=utf-8"
#        And I should receive the following health JSON response:
#        """
#            {
#                "status": "WARNING",
#                "version": {
#                    "git_commit": "3t7e5s1t4272646ef477f8ed755",
#                    "language": "go",
#                    "language_version": "go1.16.5",
#                    "version": "v1.2.3"
#                },
#                "checks": [
#                    {
#                        "name": "elasticsearch",
#                        "status": "CRITICAL",
#                        "status_code": 500,
#                        "message": "elasticsearch functionality is unavailable or non-functioning"
#                    }
#                ]
#            }
#        """
#
#    Scenario: Returning a CRITICAL (500) status when health endpoint called
#        Given one of the downstream services is failing
#        And I wait "2" seconds for the healthcheck to be available
#        When I GET "/health"
#        And I wait "3" seconds to pass the critical timeout
#        And I GET "/health"
#        Then the HTTP status code should be "500"
#        And the response header "Content-Type" should be "application/json; charset=utf-8"
#        And I should receive the following health JSON response:
#        """
#            {
#                "status": "CRITICAL",
#                "version": {
#                    "git_commit": "3t7e5s1t4272646ef477f8ed755",
#                    "language": "go",
#                    "language_version": "go1.16.5",
#                    "version": "v1.2.3"
#                },
#                "checks": [
#                    {
#                        "name": "elasticsearch",
#                        "status": "CRITICAL",
#                        "status_code": 500,
#                        "message": "elasticsearch functionality is unavailable or non-functioning"
#                    }
#                ]
#            }
#        """