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

        document.getElementById('embed-mp3-file').addEventListener('change', function() {
            app.ui.showFileInfo(this, 'embed-mp3-info', 'embed-mp3-player', 'embed-mp3-source');
        });
        document.getElementById('embed-secret-file').addEventListener('change', function() {
            app.ui.showFileInfo(this, 'embed-secret-info');
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
        
        const keyValidation = ValidationService.validateKeyForSubmission(keyInput.value);
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
    formData.set('mode', 'paper');

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
        
        const form = event.target;
        const keyInput = document.getElementById('extract-key');
        
        const keyValidation = ValidationService.validateKeyForSubmission(keyInput.value);
        if (!keyValidation.isValid) {
            this.ui.showResult('Error: ' + keyValidation.error, true);
            return;
        }

        const mp3Validation = ValidationService.validateFile(
            document.getElementById('extract-mp3-file'), 'mp3'
        );
        if (!mp3Validation.isValid) {
            this.ui.showResult('Error: ' + mp3Validation.error, true);
            return;
        }

        const formData = new FormData(form);
        formData.set('use_encryption', document.getElementById('extract-encryption').checked);
        formData.set('use_key_for_position', document.getElementById('extract-position').checked);
        formData.set('lsb_bits', document.getElementById('extract-lsb').value);
    formData.set('mode', 'paper');

        try {
            this.ui.showResult('Extracting... Please wait', false);
            
            const result = await this.api.extractFile(formData);
            
            this.ui.showExtractResult(
                'Secret file extracted successfully! Preview the result below:', 
                result.blob, 
                result.contentType
            );
        } catch (error) {
            console.error('Error extracting file:', error);
            this.ui.showResult('Error: ' + error.message, true);
        }
    }
    
    async handlePSNR(event) {
        event.preventDefault();
        
        const form = event.target;
        
        // Validate original file
        const originalFileValidation = ValidationService.validateFile(
            document.getElementById('psnr-original-file'), 'mp3'
        );
        if (!originalFileValidation.isValid) {
            this.ui.showResult('Error: ' + originalFileValidation.error, true);
            return;
        }
        
        // Validate modified file
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
}

document.addEventListener('DOMContentLoaded', function() {
    window.app = new SteganographyApp();
});