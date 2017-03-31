Feature: Searching ONS website content
  Scenario: When viewing the release calender I want to see the Published releases
    When a user filters the release calendar for "published" documents
    Then user will receive a list of the documents are published
