
# Description
This Docker image provides a microservice that exposes an endpoint for retrieving Kanban metrics for a task in ClickUp.

# Usage
To run the container, use the following command:
`docker run -it -p 8080:8080 -e API_KEY=<your_api_key> -e GITLAB_TOKEN=<your_gitlab_token> lucasvillalba/software-delivery-metrics:latest`

# Endpoint
The microservice exposes the following endpoint:
`GET /metrics/{task_id}`

## Example
To retrieve metrics for a task with ID `12345`, make a `GET` request to:
`http://localhost:8080/metrics/12345`

Make sure to include the required authorization token in the request header. Use the Authorization header with the value `Bearer {your_token}`. Replace `{your_token}` with your actual authorization token.

Here is an example using cURL:

`curl -H "Authorization: Bearer {your_token}" http://localhost:8080/metrics/12345`

This endpoint retrieves Kanban metrics for a task specified by `{task_id}` in ClickUp.

# Environment Variables
The following environment variable is required for configuring the microservice:

`API_KEY:` ClickUp API key for authentication with the ClickUp API.