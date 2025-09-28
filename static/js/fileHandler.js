export class FileHandler {
    constructor(uiManager) {
        this.ui = uiManager;
    }

    async downloadStegoFile() {
        if (!this.ui.currentStegoBlob) return;
        
        const defaultFilename = this.ui.currentOriginalFilename ? 
            'stego_' + this.ui.currentOriginalFilename : 
            'stego_audio.mp3';
        
        if ('showSaveFilePicker' in window) {
            try {
                const fileHandle = await window.showSaveFilePicker({
                    suggestedName: defaultFilename,
                    types: [{
                        description: 'Audio files',
                        accept: {
                            'audio/mpeg': ['.mp3'],
                            'audio/*': ['.mp3', '.wav', '.m4a']
                        }
                    }]
                });
                
                const writable = await fileHandle.createWritable();
                await writable.write(this.ui.currentStegoBlob);
                await writable.close();
                
                this.ui.showResult('File saved successfully!', false);
                return;
            } catch (err) {
                if (err.name !== 'AbortError') {
                    console.error('Error saving file:', err);
                }
                return;
            }
        }
        
        this._triggerDownload(this.ui.currentStegoUrl, defaultFilename);
    }

    async downloadExtractedFile() {
        if (!this.ui.currentExtractBlob) return;
        
        let defaultFilename = 'extracted_secret.txt';
        if (this.ui.currentExtractContentType) {
            const extension = this._getExtensionFromMimeType(this.ui.currentExtractContentType);
            defaultFilename = 'extracted_secret' + extension;
        }
        
        if ('showSaveFilePicker' in window) {
            try {
                const acceptTypes = this._getAcceptTypesFromMimeType(this.ui.currentExtractContentType);
                
                const fileHandle = await window.showSaveFilePicker({
                    suggestedName: defaultFilename,
                    types: [{
                        description: acceptTypes.description,
                        accept: acceptTypes.accept
                    }]
                });
                
                const writable = await fileHandle.createWritable();
                await writable.write(this.ui.currentExtractBlob);
                await writable.close();
                
                this.ui.showResult('File saved successfully!', false);
                return;
            } catch (err) {
                if (err.name !== 'AbortError') {
                    console.error('Error saving file:', err);
                }
                return;
            }
        }
        
        this._triggerDownload(this.ui.currentExtractUrl, defaultFilename);
    }

    _triggerDownload(url, filename) {
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
    }

    _getExtensionFromMimeType(mimeType) {
        const mimeToExt = {
            'text/plain': '.txt',
            'image/jpeg': '.jpg',
            'image/png': '.png',
            'image/gif': '.gif',
            'image/bmp': '.bmp',
            'application/pdf': '.pdf',
            'audio/mpeg': '.mp3',
            'audio/wav': '.wav',
            'video/mp4': '.mp4',
            'application/zip': '.zip',
            'application/json': '.json'
        };
        return mimeToExt[mimeType] || '.dat';
    }

    _getAcceptTypesFromMimeType(mimeType) {
        if (!mimeType) {
            return {
                description: 'All files',
                accept: { '*/*': ['*'] }
            };
        }

        const typeMap = {
            'text/plain': {
                description: 'Text files',
                accept: { 'text/plain': ['.txt'] }
            },
            'image/jpeg': {
                description: 'JPEG images',
                accept: { 'image/jpeg': ['.jpg', '.jpeg'] }
            },
            'image/png': {
                description: 'PNG images',
                accept: { 'image/png': ['.png'] }
            },
            'application/pdf': {
                description: 'PDF files',
                accept: { 'application/pdf': ['.pdf'] }
            }
        };

        return typeMap[mimeType] || {
            description: 'All files',
            accept: { '*/*': [this._getExtensionFromMimeType(mimeType)] }
        };
    }
}