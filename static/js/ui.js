export class UIManager {
    constructor() {
        this.currentStegoBlob = null;
        this.currentExtractBlob = null;
        this.currentStegoUrl = null;
        this.currentExtractUrl = null;
        this.currentOriginalFilename = null;
        this.currentExtractContentType = null;
    }

    showConnectionStatus(isConnected, message = '') {
        const statusDiv = document.getElementById('status');
        if (isConnected) {
            statusDiv.className = 'text-center mb-5 p-3 rounded-md font-bold border-2 border-green-200 bg-green-50 text-green-800';
            statusDiv.textContent = 'Connected to Steganography Server';
        } else {
            statusDiv.className = 'text-center mb-5 p-3 rounded-md font-bold border-2 border-red-200 bg-red-50 text-red-800';
            statusDiv.textContent = message || 'Cannot connect to server';
        }
    }

    switchTab(tab) {
        const tabs = document.querySelectorAll('.flex-1');
        
        tabs.forEach(t => {
            t.className = 'flex-1 p-4 border-2 border-gray-300 bg-white text-gray-700 rounded-lg cursor-pointer text-center transition-all duration-300 hover:border-blue-500 hover:bg-gray-50';
        });
        
        if (tab === 'embed') {
            tabs[0].className = 'flex-1 p-4 border-2 border-blue-500 bg-blue-500 text-white rounded-lg cursor-pointer text-center transition-all duration-300 hover:bg-blue-600';
        } else if (tab === 'extract') {
            tabs[1].className = 'flex-1 p-4 border-2 border-green-500 bg-green-500 text-white rounded-lg cursor-pointer text-center transition-all duration-300 hover:bg-green-600';
        } else if (tab === 'psnr') {
            tabs[2].className = 'flex-1 p-4 border-2 border-purple-500 bg-purple-500 text-white rounded-lg cursor-pointer text-center transition-all duration-300 hover:bg-purple-600';
        }

        document.getElementById('embed-section').classList.add('hidden');
        document.getElementById('extract-section').classList.add('hidden');
        document.getElementById('psnr-section').classList.add('hidden');
        
        if (tab === 'embed') {
            document.getElementById('embed-section').classList.remove('hidden');
        } else if (tab === 'extract') {
            document.getElementById('extract-section').classList.remove('hidden');
        } else if (tab === 'psnr') {
            document.getElementById('psnr-section').classList.remove('hidden');
        }

        this.hideResultSection();
    }

    hideResultSection() {
        const resultSection = document.getElementById('result-section');
        const resultPlayer = document.getElementById('result-player');
        const extractResult = document.getElementById('extract-result');
        const psnrResult = document.getElementById('psnr-result');
        const audioSource = document.getElementById('result-mp3-source');
        
        if (audioSource.src && audioSource.src.startsWith('blob:')) {
            URL.revokeObjectURL(audioSource.src);
            audioSource.src = '';
        }
        
        if (this.currentStegoUrl) {
            URL.revokeObjectURL(this.currentStegoUrl);
            this.currentStegoUrl = null;
        }
        
        if (this.currentExtractUrl) {
            URL.revokeObjectURL(this.currentExtractUrl);
            this.currentExtractUrl = null;
        }
        
        this.currentStegoBlob = null;
        this.currentExtractBlob = null;
        
        resultPlayer.classList.add('hidden');
        extractResult.classList.add('hidden');
        psnrResult.classList.add('hidden');
        resultSection.classList.add('hidden');
    }

    showResult(message, isError = false) {
        const resultSection = document.getElementById('result-section');
        const resultMessage = document.getElementById('result-message');
        const resultPlayer = document.getElementById('result-player');
        const extractResult = document.getElementById('extract-result');
        
        resultPlayer.classList.add('hidden');
        extractResult.classList.add('hidden');
        
        resultMessage.innerHTML = `
            <div class="p-4 rounded-lg font-bold border-2 ${isError ? 'border-red-200 bg-red-50 text-red-800' : 'border-green-200 bg-green-50 text-green-800'}">
                ${message}
            </div>
        `;
        
        resultSection.classList.remove('hidden');
    }

    showEmbedResult(message, audioBlob, originalFilename) {
        const resultSection = document.getElementById('result-section');
        const resultMessage = document.getElementById('result-message');
        const resultPlayer = document.getElementById('result-player');
        const extractResult = document.getElementById('extract-result');
        const audioSource = document.getElementById('result-mp3-source');
        
        extractResult.classList.add('hidden');
        resultPlayer.classList.remove('hidden');
        
        this.currentStegoBlob = audioBlob;
        this.currentOriginalFilename = originalFilename;
        if (this.currentStegoUrl) {
            URL.revokeObjectURL(this.currentStegoUrl);
        }
        this.currentStegoUrl = URL.createObjectURL(audioBlob);
        
        resultMessage.innerHTML = `
            <div class="p-4 rounded-lg font-bold border-2 border-green-200 bg-green-50 text-green-800">
                ${message}
            </div>
        `;
        
        audioSource.src = this.currentStegoUrl;
        
        const audioElement = resultPlayer.querySelector('audio');
        audioElement.load();
        
        resultSection.classList.remove('hidden');
    }

    showExtractResult(message, extractBlob, contentType) {
        const resultSection = document.getElementById('result-section');
        const resultMessage = document.getElementById('result-message');
        const resultPlayer = document.getElementById('result-player');
        const extractResult = document.getElementById('extract-result');
        const extractPreview = document.getElementById('extract-preview');
        
        resultPlayer.classList.add('hidden');
        extractResult.classList.remove('hidden');
        
        this.currentExtractBlob = extractBlob;
        this.currentExtractContentType = contentType;
        if (this.currentExtractUrl) {
            URL.revokeObjectURL(this.currentExtractUrl);
        }
        this.currentExtractUrl = URL.createObjectURL(extractBlob);
        
        resultMessage.innerHTML = `
            <div class="p-4 rounded-lg font-bold border-2 border-green-200 bg-green-50 text-green-800">
                ${message}
            </div>
        `;
        
        if (contentType && contentType.startsWith('text/')) {
            const reader = new FileReader();
            reader.onload = function(e) {
                extractPreview.innerHTML = `<pre class="whitespace-pre-wrap text-sm">${e.target.result}</pre>`;
            };
            reader.readAsText(extractBlob);
        } else {
            extractPreview.innerHTML = `
                <div class="text-center text-gray-600">
                    <p>Binary file detected</p>
                    <p class="text-sm">Size: ${(extractBlob.size / 1024).toFixed(2)} KB</p>
                    <p class="text-sm">Type: ${contentType || 'Unknown'}</p>
                </div>
            `;
        }
        
        resultSection.classList.remove('hidden');
    }

    showFileInfo(input, infoId, playerId = null, sourceId = null) {
        const file = input.files[0];
        const infoDiv = document.getElementById(infoId);
        
        if (file) {
            const size = (file.size / 1024 / 1024).toFixed(2);
            infoDiv.innerHTML = `${file.name} (${size} MB)`;
            infoDiv.classList.remove('hidden');
            
            if (playerId && sourceId && file.type === 'audio/mpeg') {
                const playerDiv = document.getElementById(playerId);
                const sourceElement = document.getElementById(sourceId);
                
                const objectURL = URL.createObjectURL(file);
                sourceElement.src = objectURL;
                
                playerDiv.classList.remove('hidden');
                
                const audioElement = playerDiv.querySelector('audio');
                audioElement.load();
            } else if (playerId) {
                const playerDiv = document.getElementById(playerId);
                playerDiv.classList.add('hidden');
            }
        } else {
            infoDiv.classList.add('hidden');
            if (playerId) {
                document.getElementById(playerId).classList.add('hidden');
            }
        }
    }

    showCapacityInfo(capacityData) {
        const capacityDiv = document.getElementById('embed-capacity-info');
        const capacityBytes = document.getElementById('capacity-bytes');
        const capacityMethod = document.getElementById('capacity-method');
        const capacityFrames = document.getElementById('capacity-frames');

        if (capacityData && capacityData.success) {
            capacityBytes.textContent = capacityData.capacity_readable;
            capacityMethod.textContent = capacityData.method;
            capacityFrames.textContent = capacityData.frame_count + ' frames';
            capacityDiv.classList.remove('hidden');
            
            this.currentCapacity = capacityData.capacity_bytes;
        } else {
            capacityDiv.classList.add('hidden');
            this.currentCapacity = null;
        }
    }

    hideCapacityInfo() {
        const capacityDiv = document.getElementById('embed-capacity-info');
        capacityDiv.classList.add('hidden');
        this.currentCapacity = null;
    }

    checkCapacityWarning(secretFileSize) {
        const warningDiv = document.getElementById('capacity-warning');
        const warningText = document.getElementById('capacity-warning-text');

        if (this.currentCapacity && secretFileSize > this.currentCapacity) {
            const sizeDiff = ((secretFileSize - this.currentCapacity) / 1024).toFixed(2);
            warningText.textContent = `The secret file (${(secretFileSize/1024).toFixed(2)} KB) is ${sizeDiff} KB larger than the MP3 capacity (${(this.currentCapacity/1024).toFixed(2)} KB).`;
            warningDiv.classList.remove('hidden');
            return false;
        } else {
            warningDiv.classList.add('hidden');
            return true;
        }
    }

    hideCapacityWarning() {
        const warningDiv = document.getElementById('capacity-warning');
        warningDiv.classList.add('hidden');
    }
    
    showPSNRResult(psnrData) {
        const resultSection = document.getElementById('result-section');
        const resultMessage = document.getElementById('result-message');
        const psnrResult = document.getElementById('psnr-result');
        const resultPlayer = document.getElementById('result-player');
        const extractResult = document.getElementById('extract-result');
        
        resultPlayer.classList.add('hidden');
        extractResult.classList.add('hidden');
        psnrResult.classList.remove('hidden');
        
        const psnrValue = psnrData.psnr.toFixed(2);
        document.getElementById('psnr-value').textContent = `${psnrValue} dB`;
        
        document.getElementById('mse-value').textContent = psnrData.mse.toFixed(6);
        
        document.getElementById('max-signal-value').textContent = psnrData.max_signal.toFixed(0);
        document.getElementById('original-size').textContent = this.formatBytes(psnrData.original_size);
        document.getElementById('modified-size').textContent = this.formatBytes(psnrData.modified_size);
        
        const qualityPercentage = Math.min(100, Math.max(0, (psnrValue / 50) * 100));
        const qualityBar = document.getElementById('psnr-quality-bar');
        qualityBar.style.width = `${qualityPercentage}%`;
        qualityBar.textContent = `${Math.round(qualityPercentage)}%`;
        
        if (psnrValue < 20) {
            qualityBar.classList.remove('bg-yellow-500', 'bg-green-500', 'bg-purple-500');
            qualityBar.classList.add('bg-red-500');
        } else if (psnrValue < 40) {
            qualityBar.classList.remove('bg-red-500', 'bg-green-500', 'bg-purple-500');
            qualityBar.classList.add('bg-yellow-500');
        } else {
            qualityBar.classList.remove('bg-red-500', 'bg-yellow-500', 'bg-purple-500');
            qualityBar.classList.add('bg-green-500');
        }
        
        let qualityMessage;
        if (psnrValue >= 40) {
            qualityMessage = 'Excellent quality! The modifications are practically invisible.';
        } else if (psnrValue >= 30) {
            qualityMessage = 'Good quality. Modifications are difficult to detect.';
        } else if (psnrValue >= 20) {
            qualityMessage = 'Acceptable quality. Some degradation might be noticeable.';
        } else {
            qualityMessage = 'Poor quality. Modifications are likely noticeable.';
        }
        
        resultMessage.innerHTML = `
            <div class="p-4 rounded-lg font-bold border-2 border-purple-200 bg-purple-50 text-purple-800">
                PSNR Analysis Complete: ${qualityMessage}
            </div>
        `;
        
        resultSection.classList.remove('hidden');
    }
    
    formatBytes(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        
        return parseFloat((bytes / Math.pow(1024, i)).toFixed(2)) + ' ' + sizes[i];
    }
}
