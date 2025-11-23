# **Test assignment for a Backend intern (Autumn Wave 2025)**

## **Service for assigning reviewers for Pull Requests**

A single microservice is required within the team, which automatically assigns reviewers to Pull Requests (PR), and also allows you to manage teams and participants. Interaction takes place exclusively through the HTTP API.

## **Task description**

It is necessary to implement a service that assigns reviewers to PR from the author's team, allows you to reassign reviewers and get a list of PRS assigned to a specific user, as well as manage teams and user activity. After merge PR, changing the composition of reviewers is prohibited.

## **General introductory**

**User (User)** — a team member with a unique identifier, name, and the `isActive' activity flag.

**Team (Team)** — a group of users with a unique name.

**Pull Request (PR)** — an entity with an identifier, name, author, `OPEN|MERGED' status, and a list of assigned reviewers (up to 2).

1. When creating a PR, ** up to two** active reviewers from the **author's team** are automatically assigned, excluding the author himself.
2. Reassignment replaces one reviewer with a random **active** participant ** from the replaced** reviewer's team.
3. After `MERGED`, the list of reviewers ** cannot be changed**.
4. If there are fewer than two available candidates, the available number (0/1) is assigned.

## **Conditions**

* Use this API (the OpenAPI specification will be provided in a separate file, `openapi.yaml`).
* The amount of data is moderate (up to 20 teams and up to 200 users), the RPS is 5, the response time SLI is 300 ms, and the success rate SLI is 99.9%.
* A user with `isActive = false` should not be assigned to the review.
* The merge operation must be **idempotent** — a repeat call does not result in an error and returns the current state of the PR.
* The service and its dependencies must be lifted by the **docker-compose up** command. If the solution provides for migrations, they should also be applied when executing this command. The service must be available on port 8080.
* Keep in mind that meeting the conditions for raising the service will speed up and simplify the review of your work by mentors.

## **Additional tasks**

These tasks are not mandatory, but completing all or part of them will give you an advantage over other candidates.

* Add a simple statistics endpoint (for example, the number of appointments by users and/or PR).
* Perform load testing of the resulting solution and attach brief test results to the solution.
* Add a method for mass deactivation of team users and secure reassignment of open PRS (aim to keep within 100 ms for average amounts of data).
* Implement integration or E2E testing.
* Describe the configuration of the linter.

## **Stack Requirements**

**The language of the service:** Go will be preferred, while you can choose any one that suits you.

**Database:** PostgreSQL will be preferred, and you can choose any one that suits you (an in-memory implementation is allowed).

## **Course of the decision**

If you have any questions about the assignment, the answers to which you will not find in the described "Conditions", you are free to make decisions on your own.  
In this case, attach a README file to the project, which will contain a list of questions and explanations about what assumptions you made and why exactly in the way you chose.

## **Making a decision**

You must provide a public git repository on any public host (GitHub/GitLab/etc) containing the master/main branch.:

1. Service code
2. Makefile with project build commands / Described in README.md launch instructions
3. Described in README.md questions/problems encountered and your logic for solving them (if required)