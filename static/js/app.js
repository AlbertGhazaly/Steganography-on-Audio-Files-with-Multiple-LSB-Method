import { ApiService } from './api.js';
import { UIManager } from './ui.js';
import { ValidationService } from './validation.js';
import { FileHandler } from './fileHandler.js';

class SteganographyApp {
    constructor() {
        this.api = new ApiService();
        this.ui = new UIManager();
        this.fileHandler = new FileHandler(this.ui);
        
        this.init();
    }

    async init() {
        await this.checkConnection();

        this.setupEventListeners();

        this.setupFormHandlers();

        this.handleMethodChange('embed', 'lsb');
        this.handleMethodChange('extract', 'lsb');
    }

    async checkConnection() {
        try {
            await this.api.checkHealth();
            this.ui.showConnectionStatus(true);
        } catch (error) {
            this.ui.showConnectionStatus(false, error.message);
            console.error('Connection error:', error);
        }
    }

    setupEventListeners() {
        window.switchTab = (tab) => {
            this.ui.switchTab(tab);
        };

        document.getElementById('embed-key').addEventListener('input', function() {
            ValidationService.validateKey(this);
        });
        document.getElementById('extract-key').addEventListener('input', function() {
            ValidationService.validateKey(this);
        });

        document.getElementById('embed-method').addEventListener('change', function() {
            app.handleMethodChange('embed', this.value);
        });
        document.getElementById('extract-method').addEventListener('change', function() {
            app.handleMethodChange('extract', this.value);
        });

        document.getElementById('embed-lsb').addEventListener('change', function() {
            const mp3FileInput = document.getElementById('embed-mp3-file');
            if (mp3FileInput.files[0]) {
                app.calculateCapacity(mp3FileInput.files[0], 'lsb');
            }
        });

        document.getElementById('embed-mp3-file').addEventListener('change', async function() {
            app.ui.showFileInfo(this, 'embed-mp3-info', 'embed-mp3-player', 'embed-mp3-source');
            
            if (this.files[0]) {
                await app.calculateCapacity(this.files[0]);
            } else {
                app.ui.hideCapacityInfo();
            }
        });
        document.getElementById('embed-secret-file').addEventListener('change', function() {
            app.ui.showFileInfo(this, 'embed-secret-info');
            
            if (this.files[0]) {
                app.ui.checkCapacityWarning(this.files[0].size);
            } else {
                app.ui.hideCapacityWarning();
            }
        });
        document.getElementById('extract-mp3-file').addEventListener('change', function() {
            app.ui.showFileInfo(this, 'extract-mp3-info', 'extract-mp3-player', 'extract-mp3-source');
        });
        document.getElementById('psnr-original-file').addEventListener('change', function() {
            app.ui.showFileInfo(this, 'psnr-original-info');
        });
        document.getElementById('psnr-modified-file').addEventListener('change', function() {
            app.ui.showFileInfo(this, 'psnr-modified-info');
        });

        window.downloadStegoFile = () => {
            this.fileHandler.downloadStegoFile();
        };
        window.downloadExtractedFile = () => {
            this.fileHandler.downloadExtractedFile();
        };
    }

    setupFormHandlers() {
        document.getElementById('embedForm').addEventListener('submit', (event) => {
            this.handleEmbed(event);
        });

        document.getElementById('extractForm').addEventListener('submit', (event) => {
            this.handleExtract(event);
        });
        
        document.getElementById('psnrForm').addEventListener('submit', (event) => {
            this.handlePSNR(event);
        });
    }

    async handleEmbed(event) {
        event.preventDefault();
        
        const form = event.target;
        const keyInput = document.getElementById('embed-key');
        
        const keyValidation = ValidationService.validateKeyForSubmissionRequired(keyInput.value);
        if (!keyValidation.isValid) {
            this.ui.showResult('Error: ' + keyValidation.error, true);
            return;
        }

        const mp3Validation = ValidationService.validateFile(
            document.getElementById('embed-mp3-file'), 'mp3'
        );
        if (!mp3Validation.isValid) {
            this.ui.showResult('Error: ' + mp3Validation.error, true);
            return;
        }

        const secretValidation = ValidationService.validateFile(
            document.getElementById('embed-secret-file')
        );
        if (!secretValidation.isValid) {
            this.ui.showResult('Error: ' + secretValidation.error, true);
            return;
        }

        const formData = new FormData(form);
        formData.set('use_encryption', document.getElementById('embed-encryption').checked);
        formData.set('use_key_for_position', document.getElementById('embed-position').checked);
        formData.set('lsb_bits', document.getElementById('embed-lsb').value);

        try {
            this.ui.showResult('Processing... Please wait', false);
            
            const result = await this.api.embedFile(formData);
            const originalFilename = document.getElementById('embed-mp3-file').files[0].name;
            
            this.ui.showEmbedResult(
                'Secret file embedded successfully! Preview the result below:', 
                result.blob, 
                originalFilename
            );
        } catch (error) {
            console.error('Error embedding file:', error);
            this.ui.showResult('Error: ' + error.message, true);
        }
    }

    async handleExtract(event) {
        event.preventDefault();
        console.log('Extract form submitted');
        
        const form = event.target;
        const keyInput = document.getElementById('extract-key');
        const method = document.getElementById('extract-method').value;
        
        console.log('Method:', method);
        console.log('Key value:', keyInput.value);
        
        if (keyInput.value.trim() !== '') {
            console.log('Validating key...');
            const keyValidation = ValidationService.validateKeyForSubmission(keyInput.value);
            if (!keyValidation.isValid) {
                console.error('Key validation failed:', keyValidation.error);
                this.ui.showResult('Error: ' + keyValidation.error, true);
                return;
            }
            console.log('Key validation passed');
        } else {
            console.log('No key provided, skipping validation');
        }

        console.log('Validating MP3 file...');
        const mp3Validation = ValidationService.validateFile(
            document.getElementById('extract-mp3-file'), 'mp3'
        );
        if (!mp3Validation.isValid) {
            console.error('MP3 validation failed:', mp3Validation.error);
            this.ui.showResult('Error: ' + mp3Validation.error, true);
            return;
        }
        console.log('MP3 validation passed');

        const formData = new FormData(form);
        
        if (method === 'lsb') {
            formData.delete('use_encryption');
            formData.delete('use_key_for_position');
            formData.delete('lsb_bits');
        }        try {
            console.log('Starting extraction...');
            this.ui.showResult('Extracting... Please wait', false);
            
            console.log('Calling API...');
            const result = await this.api.extractFile(formData);
            console.log('API call successful, result:', result);
            
            this.ui.showExtractResult(
                'Secret file extracted successfully! Preview the result below:', 
                result.blob, 
                result.contentType,
                result.metadata
            );
        } catch (error) {
            console.error('Error extracting file:', error);
            this.ui.showResult('Error: ' + error.message, true);
        }
    }

    async calculateCapacity(mp3File, method = null) {
        try {
            const formData = new FormData();
            formData.append('mp3_file', mp3File);
            
            if (!method) {
                const methodSelect = document.getElementById('embed-method');
                method = methodSelect ? methodSelect.value : 'header';
            }
            formData.append('method', method);
            
            if (method === 'lsb') {
                const lsbSelect = document.getElementById('embed-lsb');
                const lsbBits = lsbSelect ? lsbSelect.value : '1';
                formData.append('lsb_bits', lsbBits);
            }
            
            const capacityData = await this.api.getCapacity(formData);
            this.ui.showCapacityInfo(capacityData);
            
            const secretFileInput = document.getElementById('embed-secret-file');
            if (secretFileInput.files[0]) {
                this.ui.checkCapacityWarning(secretFileInput.files[0].size);
            }
        } catch (error) {
            console.error('Error calculating capacity:', error);
            this.ui.hideCapacityInfo();
        }
    }

    async handlePSNR(event) {
        event.preventDefault();
        
        const form = event.target;
        
        const originalFileValidation = ValidationService.validateFile(
            document.getElementById('psnr-original-file'), 'mp3'
        );
        if (!originalFileValidation.isValid) {
            this.ui.showResult('Error: ' + originalFileValidation.error, true);
            return;
        }
        
        const modifiedFileValidation = ValidationService.validateFile(
            document.getElementById('psnr-modified-file'), 'mp3'
        );
        if (!modifiedFileValidation.isValid) {
            this.ui.showResult('Error: ' + modifiedFileValidation.error, true);
            return;
        }
        
        const formData = new FormData(form);
        
        try {
            this.ui.showResult('Calculating PSNR... Please wait', false);
            
            const result = await this.api.calculatePSNR(formData);
            
            this.ui.showPSNRResult(result);
        } catch (error) {
            console.error('Error calculating PSNR:', error);
            this.ui.showResult('Error: ' + error.message, true);
        }
    }

    handleMethodChange(section, method) {
        const lsbSection = document.getElementById(`${section}-lsb-section`);
        const lsbOptions = document.getElementById(`${section}-lsb-options`);
        const keyInput = document.getElementById(`${section}-key`);
        const keyRequiredIndicator = document.getElementById(`${section}-key-required-indicator`);
        
        if (method === 'lsb') {
            if (section === 'embed') {
                lsbSection?.classList.remove('hidden');
                lsbOptions?.classList.remove('hidden');
                
                keyInput.required = true;
                if (keyRequiredIndicator) {
                    keyRequiredIndicator.textContent = '(Required for LSB)';
                    keyRequiredIndicator.style.color = '#dc2626';
                }
                
                const description = document.getElementById('method-description');
                if (description) {
                    description.textContent = 'LSB method embeds data by modifying the least significant bits of audio samples - offers more capacity but less stealth.';
                }
            } else {
                keyInput.required = false;
                if (keyRequiredIndicator) {
                    keyRequiredIndicator.textContent = '(Optional - only if encryption was used)';
                    keyRequiredIndicator.style.color = '#6b7280';
                }
            }
        } else {
            lsbSection?.classList.add('hidden');
            lsbOptions?.classList.add('hidden');
            
            keyInput.required = false;
            if (keyRequiredIndicator) {
                keyRequiredIndicator.textContent = '(Optional for Header)';
                keyRequiredIndicator.style.color = '#6b7280';
            }
            
            if (section === 'embed') {
                const description = document.getElementById('method-description');
                if (description) {
                    description.textContent = 'Header method embeds data in MP3 frame headers - more stealthy and robust.';
                }
            }
        }

        if (section === 'embed') {
            const mp3FileInput = document.getElementById('embed-mp3-file');
            if (mp3FileInput.files[0]) {
                this.calculateCapacity(mp3FileInput.files[0], method);
            }
        }
    }
}

document.addEventListener('DOMContentLoaded', function() {
    window.app = new SteganographyApp();
});