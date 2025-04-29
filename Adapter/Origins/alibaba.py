import os
from openai import OpenAI

client = OpenAI(
    # If the environment variable is not configured, replace the following line with: api_key="sk-xxx",
    api_key="sk-d199dc789f9d45d7b29a459eb2169c35",
    base_url="https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
)
completion = client.chat.completions.create(
    model="qwen-vl-plus",
    messages=[{"role": "user","content": [
            {"type": "text","text": "What is this"},
            {"type": "image_url",
             "image_url": {"url": "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg"}}
            ]}]
    )
print(completion.model_dump_json())