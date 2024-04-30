# Kahoot Clone

## Table of contents
1. [Tech stack](#tech-stack)
2. [Roles and Responsibilities](#roles-and-responsibilities)
3. [Collaboration](#collaboration)
4. [Next steps](#next-steps)

## Tech stack
This is a list of technologies that we will use in the project:

- Backend: Golang
- OpenAPI for defining the HTTP endpoints
- Golang test framework for unit tests
- HTMX for frontend interactions 
- Golang builtin template engine for rendering the pages
- Tailwind CSS for styling
- Websockets for the communication inside the game session
- Google Cloud for hosting
- Postgres for the database 
    - relational database will allow easier implementation of advanced features like stats
    - scalability of nosql is not an issue as we wouldn't use mongodb cluster in this project 
- Github for hosting the git repository
- Github actions for CI/CD
- Github project board for tracking the progress
- Github issues for communication + tracking the progress
- Docker for containerization
- Docker compose for orchestration (backend, database, frontend)
- Cypress/playwright for e2e tests

### Considerations
- Swagger for generating the API client code for interaction between frontend and backend (not needed at the moment because of HTMX) 
- Linear for project management (if github project board is not enough)

## Roles and Responsibilities
### Tasks for the whole team
- Define game rules
    - how many points for the correct answer
    - how many points for fast answer
    - do the order of the answers from the player influence the points distribution
    - how many answers can a question have
    - what types of media can be used in the questions and answers (images, videos, LaTex etc.)
    - how long is the timer for the question
	- is it the same for all questions? 
	- can it be changed by the teacher?
- Define the game flow
    - Teacher creates a lobby
    - Users join the lobby
    - Teacher starts the game
    - Questions are displayed
    - Users answer the questions
    - Results are displayed
    - Final results are displayed
- Define user stories
    - e.x. as a player I want to be able to join a lobby so that I can play with my friends
    - e.x. as a player I want to be able to create a quiz so that I can use it multiple times

### Backend
The backend team will need to do the following:

- Define API endpoints using openAPI (create lobby, join lobby, game stats, login, register etc.)
- Implement HTTP API endpoints for the defined openAPI
- Create a schema for the database
- Setup Postgres database
- Implement the game session logic and interactions through websockets
- Create unit tests for the implemented features 
    (helps during the development, makes code more reliable + shows how to use the code to others on the team)

### Frontend + UX
The frontend team will need to do the following:

- Design the UI/UX of the game, it should be responsive and easy to use
- Design the UI/UX of the website, it should be responsive and easy to use
- Implement the game UI
- Implement the website UI 
    - create quizzes
    - create lobbies
    - join lobbies
    - browse lobbies
    - browse quizzes
    - browse past games
    - etc.
- *Create e2e tests for user stories using cypress or playwright*

Considerations for the future:

- login page
- register page

### Project Management
The project manager will be responsible for the following:

- Create the project board
- Organize Meetings Structure
- Select Issues for the next goal (milestone)
- Review the project diaries
- Review the pull requests
- Do the presentations
- Help team members with the issues

### Infrastructure
The infrastructure team will be responsible for the following:

- Research google cloud, and how to deploy docker containers on it
    - check how to run docker compose as well
- Setup Automatic deployment of "main" branch to the google cloud 
- Setup the CI/CD pipeline (github actions)
    - for running the tests automatically
    - for deploying the application to the google cloud
- Maintain the infrastructure
    - check if the application is running
    - check if the database is running
    - check if the tests are passing

### Team roles and Responsibilities
- Minh - backend + frontend
- Eren - backend
- Daniel - frontend + UX
- Sunny - frontend + UX
- Micmi - frontend + UX
- Elias - infrastructre
- Eryk - project manager + backend + where help is needed

## Collaboration
- The project will be hosted on github. 
- We will use issues as a way to track the progress of the project (think of issues as todo tasks) + 
    discuss the possible solutions to the problems so that the person trying to solve the issue can have a better understanding of the problem.
- Which group of issues (tasks) will be tackled next depends on the project manager
- Automatic project board to manage the tasks (issues) in the project. Think of it as visual representation of which issues are "in backlog",
    "being worked on", "waiting for merge", "done" (The issues will be automatically moved based on labels)
- Each issues will have their own branch. The branch will be named after the issue number.
    It will be marged through a pull request so that it can be reviewed.

### Meetings
The goal of the meetings is to keep everyone on the same page and to discuss the major issues that are blocking the progress.

During weekly meetings we will:

- Discuss the major issues that were solved during the week
- Discuss the major issues that are blocking the progress
- Discuss which major issues will be tackled next week

Additionaly we need to do at least one only frontend and only backend meeting so that the teams may plan the work. 
(Preferably we should do it this week)

## Next Steps
### Tasks for the week 1, 29-04-2024 - 06-05-2024
Whole team:

- Create issues (tasks) for the project and discuss their solutions inside so that when a person implementing the feature
    ,he/she has all considerations be in one place
- Familiarize with the tech stack primarly websockets
- Help to establish vacabulary (session, lobby, screen/page/site, quiz, question, answer, player, game, etc.) so that we can communicate better
- Write the project diary (what was done, what was learned)

Backend:
- Define the API endpoints using openAPI
- Design the database schema

Frontend:
- UI design prototype for the website (can be just a sketch, doesn't need to be code) by frontend team
- UI design prototype for the game (can be just a sketch, doesn't need to be code) by frontend team

PM:

- have github repo ready
    - have the project structure ready
    - golang directory structure
    - dockerfile file for backend
- have project board ready
- assign issues to the team members
- create vocabulary for the project

## Our first milestone (in 2-3 weeks), MVP
Create MVP (minimum viable product) for the project with following functionality:

- lobby/session creation
- playing the quiz with only text questions and answers
- display of results

CI/CD pipeline should be setup with testing and deployment to google cloud
Should be deployed on google cloud inside a docker container

