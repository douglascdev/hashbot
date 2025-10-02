// src/routes/your-page/+page.server.js
export const load = async ({ request }) => {
  const acceptLanguage = request.headers.get('accept-language');
  const response = await fetch('http://localhost:8080/api/commands', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Accept-Language': acceptLanguage || 'en-US', // Default to 'en-US' if not provided
        }
    });
  const commands = await response.json();
  return {
    commands
  };
};
