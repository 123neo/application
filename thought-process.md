## Running the Application

Please refer to the `Readme.md` file for detailed instructions on how to run the application.

---

## Thought Process

### Objective
Build an endpoint to capture unique requests by ID every minute and log them into a `testlogfile`.

### Approach

1. **Uniqueness Handling**:
    - Use the request ID and cookies to determine if the user is new or returning.
    - Check the uniqueness of the request ID using Redis cache:
      - If the ID exists in Redis, return the existing cookie.
      - If the ID does not exist, create a new cookie and store it in Redis.

2. **Distributed Locking**:
    - Implement distributed locking using Redis to handle concurrent requests across multiple server instances.
    - Ensure that only one instance processes a request at a time to maintain consistency.

3. **Logging Unique Requests**:
    - Every minute:
      - Retrieve all unique request IDs from Redis.
      - Log these IDs into the `testlogfile` located in the root directory.
      - Clear the Redis cache to prepare for the next cycle.

4. **Future Enhancements**:
    - Plan to push unique request data to a distributed streaming service like Kafka.
    - A Kafka container is already set up in Docker and can be started using the `docker-compose.yaml` file.
    - Due to time constraints, Kafka integration is not yet implemented. To complete this, we add code to produce messages to Kafka where we log data in `testlogfile`.

---