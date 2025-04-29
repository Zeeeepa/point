import os
from openai import OpenAI

# Initialize the OpenAI client
client = OpenAI(
    api_key=os.getenv("OPENAI_API_KEY", "your-api-key-here"),  # Use your actual API key
    base_url="http://localhost:8000"  # Ensure the base URL is correct
)

# Make a request to the OpenAI API
completion = client.chat.completions.create(
    model="gpt-3.5-turbo",  # Replace with a valid model name
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What is this"}
            ]
        }
    ]
)

# Print the response
print(completion.model_dump_json())
