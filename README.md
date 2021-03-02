# ShakeSearch

Welcome to the Pulley Shakesearch Take-home Challenge! In this repository,
you'll find a simple web app that allows a user to search for a text string in
the complete works of Shakespeare.

You can see a live version of the app at
https://pulley-shakesearch.herokuapp.com/. Try searching for "Hamlet" to display
a set of results.

In it's current state, however, the app is just a rough prototype. The search is
case sensitive, the results are difficult to read, and the search is limited to
exact matches.

## Your Mission

Improve the search backend. Think about the problem from the **user's perspective**
and prioritize your changes according to what you think is most useful. 

## Evaluation

We will be primarily evaluating based on how well the search works for users. A search result with a lot of features (i.e. multi-words and mis-spellings handled), but with results that are hard to read would not be a strong submission. 


## Submission

1. Fork this repository and send us a link to your fork after pushing your changes. 
2. Heroku hosting - The project includes a Heroku Procfile and, in its
current state, can be deployed easily on Heroku's free tier.
3. In your submission, share with us what changes you made and how you would prioritize changes if you had more time.


# Submission from Santiago Fulladoza

### What does this version include:

- With the purpose of improve the matches, we keep in the struct a version of completeWorks which is lower case and removes some symbols like dots, comma, semicolon, etc.
- Also, we keep a reference to the start and end for each paragraph in the completeWorks text, so whenever we make a search we can identify to which paragraph the query belongs (this is useful for us because of two reasons we will mention in continuation)
- We have two search modes: exact phrase and all words
     - in the exact phrase mode, we search for the entire query text, what will have successful matches only if all the words in the query are together in the text
     - in the all words mode, we search individually for each word in the query, and we will have a match for each paragraph containing all words (we use the pre-loaded paragraph start and end indexes here, and intersect the result sets of each individual search)
- For the response, we return the content of paragraph to which the queried text belongs, with a cap of 10 lines (because we don't want too large responses). A con of this is that, for long paragraphs and when working in the all words search mode, it is possible that not all the words in the query are included in the response being displayed. Another con is that sometimes we have very short responses (for 1 or 2 line paragraphs).
- Also, we identified in load time the reference for the start and end of each title in the CompleteWorks, so we also can identify to which work does the queried text belong. We are using this to display the work title for each response

### What does not include

If I have had more time,

- I would have identified, the same way I did with work titles, the ACTs and SCENEs on each work, so we can have a better reference for the found text on each response.
- I would have implemented pagination for the response as, for queries that are too open, the result list is too long.
- I would have improved the response for each finding, returning the number of the line and the index within it containing each query word, so we can set a special style for it (like a highlight) to help the user find the search words within each result.
- I would have explored using some library (or implemented some solution) for fuzzy search, so we can not just find exact terms, but similar words with some tolerance parameter. The risk of this is that the search might become much slower, as the matches can highly increase and also the search is less efficient than for suffix arrays.
- I would have added more tests. A lot more tests.

### Try the solution

You can see the live version at
https://shakesearch-fulla.herokuapp.com.

