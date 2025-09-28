export class ApiService {
    constructor() {
        this.BASE_URL = '/api';
    }

    async checkHealth() {
        try {
            const response = await fetch(`${this.BASE_URL}/health`);
            return await response.json();
        } catch (error) {
            throw new Error(`Connection failed: ${error.message}`);
        }
    }

    async embedFile(formData) {
        try {
            const response = await fetch(`${this.BASE_URL}/embed`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Embedding failed');
            }

            return {
                blob: await response.blob(),
                contentType: response.headers.get('content-type')
            };
        } catch (error) {
            throw new Error(`Embed operation failed: ${error.message}`);
        }
    }

    async extractFile(formData) {
        try {
            const response = await fetch(`${this.BASE_URL}/extract`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Extraction failed');
            }

            return {
                blob: await response.blob(),
                contentType: response.headers.get('content-type')
            };
        } catch (error) {
            throw new Error(`Extract operation failed: ${error.message}`);
        }
    }

    async getCapacity(formData) {
        try {
            const response = await fetch(`${this.BASE_URL}/capacity`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Capacity calculation failed');
            }

            return await response.json();
        } catch (error) {
            throw new Error(`Capacity calculation failed: ${error.message}`);
        }
    }
}
