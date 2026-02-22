import { Octokit } from "@octokit/core";
import { paginateGraphQL } from "@octokit/plugin-paginate-graphql";

const token = process.env.GITHUB_TOKEN;
const repository = process.env.GITHUB_REPOSITORY;
const prNumber = parseInt(process.env.PR_NUMBER, 10);

if (!token || !repository || !prNumber) {
  console.error(
    "Missing required env vars: GITHUB_TOKEN, GITHUB_REPOSITORY, PR_NUMBER"
  );
  process.exit(0);
}

const [owner, repo] = repository.split("/");

const PaginatedOctokit = Octokit.plugin(paginateGraphQL);
const octokit = new PaginatedOctokit({ auth: token });

async function fetchReviews() {
  const { repository } = await octokit.graphql.paginate(
    `
    query ($owner: String!, $repo: String!, $prNumber: Int!, $cursor: String) {
      repository(owner: $owner, name: $repo) {
        pullRequest(number: $prNumber) {
          reviews(first: 100, after: $cursor) {
            nodes {
              id
              author { login }
              isMinimized
            }
            pageInfo { hasNextPage endCursor }
          }
        }
      }
    }
  `,
    { owner, repo, prNumber }
  );

  return repository.pullRequest.reviews.nodes;
}

async function fetchIssueComments() {
  const { repository } = await octokit.graphql.paginate(
    `
    query ($owner: String!, $repo: String!, $prNumber: Int!, $cursor: String) {
      repository(owner: $owner, name: $repo) {
        pullRequest(number: $prNumber) {
          comments(first: 100, after: $cursor) {
            nodes {
              id
              author { login }
              body
              isMinimized
            }
            pageInfo { hasNextPage endCursor }
          }
        }
      }
    }
  `,
    { owner, repo, prNumber }
  );

  return repository.pullRequest.comments.nodes;
}

async function minimizeComment(subjectId) {
  await octokit.graphql(
    `
    mutation ($id: ID!) {
      minimizeComment(input: { subjectId: $id, classifier: OUTDATED }) {
        minimizedComment { isMinimized }
      }
    }
  `,
    { id: subjectId }
  );
}

try {
  console.log(`Fetching reviews and comments for PR #${prNumber}...`);

  const [reviews, comments] = await Promise.all([
    fetchReviews(),
    fetchIssueComments(),
  ]);

  const reviewsToMinimize = reviews.filter(
    (r) => r.author?.login === "claude[bot]" && !r.isMinimized
  );

  const commentsToMinimize = comments.filter(
    (c) => !c.isMinimized && c.body?.includes("/request-review")
  );

  console.log(
    `Found ${reviewsToMinimize.length} Claude review(s) and ${commentsToMinimize.length} /request-review comment(s) to minimize.`
  );

  for (const review of reviewsToMinimize) {
    console.log(`Minimizing Claude review: ${review.id}`);
    await minimizeComment(review.id);
  }

  for (const comment of commentsToMinimize) {
    console.log(
      `Minimizing /request-review comment by ${comment.author?.login}: ${comment.id}`
    );
    await minimizeComment(comment.id);
  }

  console.log("Done.");
} catch (error) {
  console.error("Error minimizing comments:", error.message);
  process.exit(0);
}
