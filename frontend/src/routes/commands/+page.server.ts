// src/routes/your-page/+page.server.js
export const load = async ({ fetch }) => {
  const response = await fetch('http://localhost:8080/api/commands');
  const commands = await response.json();
  return {
    commands
  };
};
