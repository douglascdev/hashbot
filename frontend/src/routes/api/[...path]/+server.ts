import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params, url, request }) => {
    // Construct the URL for the Go backend
    const backendUrl = `http://localhost:8080/api/${params.path}${url.search}`;

	console.log(`Forwarding request to: ${backendUrl}`);

    // Extract the Accept-Language header from the incoming request
    const acceptLanguage = request.headers.get('accept-language');
    console.log(`redirected api Accept-Language=${acceptLanguage}`)

    // Forward the request to the Go backend
    const response = await fetch(backendUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Accept-Language': acceptLanguage || 'en-US', // Default to 'en-US' if not provided
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
            'Content-Language': contentLanguage || 'en-US', // Default to 'en-US' if not provided
        }
    });
};
