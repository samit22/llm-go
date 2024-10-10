package ragserver

type GraphQLResponse struct {
	Get struct {
		Document []struct {
			Text string `json:"text"`
		} `json:"Document"`
	} `json:"Get"`
}

const Template = `
### Question:
%s

### Context:
%s
### Instructions:
- Provide a clear and concise response based on the context provided.
- Stay focused on the context and avoid making assumptions beyond the given data.
- Use the context to guide your response and provide a well-reasoned answer.
- Ensure that your response is relevant and addresses the question asked.
- If the question does not relate to the context, answer it as normal.
`
