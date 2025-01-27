## Real GraphQL Server Example

Let's create a real-world example that provides a simple GraphQL API for the non-existent Front-end.
The requirements for the project will be described later. Now, let's say that we want to create an API for the blog.
It is a simple API, but it is enough to show how to create a GraphQL server using Modulus framework and the `gqlgen` module.
Previously we created a server that has only one query `ping` that returns the string `pong`.

Read the [Getting Started](./getting_started.md) guide to create a new project with the base set of modules.

If you don't want to read this and follow all the steps from the article, please use the following commands to create a new project:

```bash
	go install github.com/go-modulus/modulus/cmd/mtools@latest
	mkdir blog
    cd blog
	mtools init --name=blog
	mtools module install -m "pgx, dbmate migrator, chi http, gqlgen"
	mtools module create --silent --path=internal --package=blog
```

### Requirements

Every project starts with requirements. Let's define the requirements for our blog API:
1. There are 2 roles in the system: `Admin` and `User`.
2. The `Admin` can create, update, and delete any posts in the system. Also, the `Admin` can see all posts, including unpublished ones.
3. The `User` can create, update, and delete only his posts and see all published posts.
4. The system shows the list of posts with the following fields: `ID`, `Title`, `Preview`, `Author`, `Status`, `PublishedAt`.
5. The system allows filtering posts by `Status` and sorting by `PublishedAt`.
6. The system allows getting the full post by `ID` providing the additional field `Content`.
7. To work with a system the users should register themselves in the system.
8. The system should provide the `login` mutation to authenticate the user.
9. The system should provide the `me` query to get the current user.
10. The system should provide the `logout` mutation to log out the user.
11. The system should provide the `refresh` mutation to refresh the user's token.
12. The `Admin` can see the list of all users in the system.
13. The `Admin` can change the role of the user.