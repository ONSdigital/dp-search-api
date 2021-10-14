#Feature: Searching ONS website content
#  Scenario: When Searching for Retail Price Index I get the same results as RPI
#    When a user searches for the term(s) "Retail Price Index"
#    Then the user will receive the first page with documents only from this uri prefix /economy/inflationandpriceindices/bulletins/consumerpriceinflation
#    And the results with the same score are in date descending order
#
#  Scenario: When Searching for RPI I get the same results as Retail Price Index
#    When a user searches for the term(s) "RPI"
#    Then the user will receive the first page with documents only from this uri prefix /economy/inflationandpriceindices/bulletins/consumerpriceinflation
#    And the results with the same score are in date descending order
#
#  Scenario: When Searching for CPI I get the same results as Consumer Price Index
#    When a user searches for the term(s) "CPI"
#    Then the user will receive the first page with documents only from this uri prefix /economy/inflationandpriceindices/bulletins/consumerpriceinflation
#    And the results with the same score are in date descending order
#
#  Scenario: When Searching for Consumer Price Index I get the same results as Consumer Price Index
#    When a user searches for the term(s) "Consumer Price Inflation"
#    Then the user will receive the first page with documents only from this uri prefix /economy/inflationandpriceindices/bulletins/consumerpriceinflation
#    And the results with the same score are in date descending order
