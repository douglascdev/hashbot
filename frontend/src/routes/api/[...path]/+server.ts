import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params, url }) => {
    // Construct the URL for the Go backend
    const backendUrl = `http://localhost:8080/api/${params.path}${url.search}`;

	console.log(`Forwarding request to: ${backendUrl}`);

    // Forward the request to the Go backend
    const response = await fetch(backendUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            // Add any other headers you need to pass to the Go backend
        }
    });

    // Check if the response is okay
    if (!response.ok) {
        return new Response('Error fetching data from backend', { status: response.status });
    }

    // Return the response from the Go backend
    const data = await response.json();
    return new Response(JSON.stringify(data), {
        status: response.status,
        headers: {
            'Content-Type': 'application/json',
        }
    });
};
