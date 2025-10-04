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

            const metadata = {
                originalFilename: response.headers.get('X-Original-Filename'),
                fileType: response.headers.get('X-File-Type'),
                secretSize: response.headers.get('X-Secret-Size'),
                usedEncryption: response.headers.get('X-Used-Encryption') === 'true',
                usedKeyPosition: response.headers.get('X-Used-Key-Position') === 'true',
                lsbBits: response.headers.get('X-LSB-Bits')
            };

            return {
                blob: await response.blob(),
                contentType: response.headers.get('content-type'),
                metadata: metadata
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

    async calculatePSNR(formData) {
        try {
            const response = await fetch(`${this.BASE_URL}/psnr`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'PSNR calculation failed');
            }

            return await response.json();
        } catch (error) {
            
            throw new Error(`PSNR calculation failed: ${error.message}`);
        }
    }
}
