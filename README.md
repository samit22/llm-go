# RAG server in Go
Example of a rag server built using go.


### Running the Application
   - Clone the application
   ```
      git clone git@github.com:samit22/llm-go.git
   ```

#### Using docker
   - Make sure docker is installed
   - Run the command
   ```
      GEMINI_FLASH_API_KEY="your-key" make start-docker
   ```

#### Starting local

##### Requirements
 - Docker to run vector DB locally
    - Installation (guide)[https://docs.docker.com/engine/install]

 - go 1.23
    - Installation (guide)[https://go.dev/doc/install]

 - API Key for Gemini
    - Follow this [link](https://aistudio.google.com/app/) to create a key

 - Export the key
    ```
     export GEMINI_FLASH_API_KEY={API_KEY}
    ```
    or create a filename called `.env.gemini-flash-api-key` and paste the value

#### Run
 - Start Vector DB
    ```bash
      docker compose up -d -f docker-compose-vector-db.yaml
    ```

 - Run the program
    ```bash
    go run .
    ```

#### Sample document and queries
- Add documents
```bash
curl --location 'localhost:5000/add-documents' \
--header 'Content-Type: application/json' \
--data '{
    "documents": ["Your name is Mastermind.", "You live in Kings Beach, Australia.", "You are developed while learning RAG server based on Go for LLMs. "]
}'
```

- Ask question
``` bash
   curl --location 'localhost:5000/ask' \
   --header 'Content-Type: application/json' \
   --data '{
      "question": "Where do you live?"
   }'
```





