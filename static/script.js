const API_BASE = '/api';

async function checkConnection() {
    try {
        const response = await fetch(`${API_BASE}/health`);
        const data = await response.json();
        
        const statusDiv = document.getElementById('status');
        if (response.ok) {
            statusDiv.className = 'text-center mb-5 p-3 rounded-md font-bold border-2 border-green-200 bg-green-50 text-green-800';
            statusDiv.textContent = 'Connected to Steganography Server';
        } else {
            throw new Error('API not responding');
        }
    } catch (error) {
        const statusDiv = document.getElementById('status');
        statusDiv.className = 'text-center mb-5 p-3 rounded-md font-bold border-2 border-red-200 bg-red-50 text-red-800';
        statusDiv.textContent = 'Cannot connect to server';
        console.error('Connection error:', error);
    }
}

function switchTab(tab) {
    const tabs = document.querySelectorAll('.flex-1');
    
    tabs.forEach(t => {
        t.className = 'flex-1 p-4 border-2 border-gray-300 bg-white text-gray-700 rounded-lg cursor-pointer text-center transition-all duration-300 hover:border-blue-500 hover:bg-gray-50';
    });
    
    if (tab === 'embed') {
        tabs[0].className = 'flex-1 p-4 border-2 border-blue-500 bg-blue-500 text-white rounded-lg cursor-pointer text-center transition-all duration-300 hover:bg-blue-600';
    } else {
        tabs[1].className = 'flex-1 p-4 border-2 border-green-500 bg-green-500 text-white rounded-lg cursor-pointer text-center transition-all duration-300 hover:bg-green-600';
    }

    if (tab === 'embed') {
        document.getElementById('embed-section').classList.remove('hidden');
        document.getElementById('extract-section').classList.add('hidden');
    } else {
        document.getElementById('embed-section').classList.add('hidden');
        document.getElementById('extract-section').classList.remove('hidden');
    }

    const resultSection = document.getElementById('result-section');
    const resultPlayer = document.getElementById('result-player');
    const extractResult = document.getElementById('extract-result');
    const audioSource = document.getElementById('result-mp3-source');
    
    if (audioSource.src && audioSource.src.startsWith('blob:')) {
        URL.revokeObjectURL(audioSource.src);
        audioSource.src = '';
    }
    
    if (currentStegoUrl) {
        URL.revokeObjectURL(currentStegoUrl);
        currentStegoUrl = null;
    }
    
    if (currentExtractUrl) {
        URL.revokeObjectURL(currentExtractUrl);
        currentExtractUrl = null;
    }
    
    currentStegoBlob = null;
    currentExtractBlob = null;
    currentOriginalFilename = null;
    currentExtractContentType = null;
    
    resultPlayer.classList.add('hidden');
    extractResult.classList.add('hidden');
    resultSection.classList.add('hidden');
}

function validateKey(input) {
    let value = input.value;
    let hasInvalidChars = false;
    
    let filteredValue = '';
    for (let i = 0; i < value.length; i++) {
        const charCode = value.charCodeAt(i);
        if (charCode <= 255) {
            filteredValue += value.charAt(i);
        } else {
            hasInvalidChars = true;
        }
    }
    
    input.value = filteredValue;
    
    const parent = input.parentElement;
    const existingError = parent.querySelector('.key-error');
    
    if (hasInvalidChars) {
        if (!existingError) {
            const errorDiv = document.createElement('div');
            errorDiv.className = 'key-error text-red-500 text-sm mt-1';
            errorDiv.textContent = 'Key can only contain ASCII characters (0-255)';
            parent.appendChild(errorDiv);
        }
    } else if (existingError) {
        existingError.remove();
    }
}

function showFileInfo(input, infoId, playerId = null, sourceId = null) {
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
            const playerDiv = document.getElementById(playerId);
            playerDiv.classList.add('hidden');
        }
    }
}

function showResult(message, isError = false) {
    const resultSection = document.getElementById('result-section');
    const resultMessage = document.getElementById('result-message');
    const resultPlayer = document.getElementById('result-player');
    
    resultMessage.innerHTML = `
        <div class="p-4 rounded-lg font-bold border-2 ${isError ? 'border-red-200 bg-red-50 text-red-800' : 'border-green-200 bg-green-50 text-green-800'}">
            ${message}
        </div>
    `;
    
    resultPlayer.classList.add('hidden');
    resultSection.classList.remove('hidden');
}

let currentStegoBlob = null;
let currentExtractBlob = null;
let currentStegoUrl = null;
let currentExtractUrl = null;
let currentOriginalFilename = null;
let currentExtractContentType = null;

function showEmbedResult(message, audioBlob, originalFilename) {
    const resultSection = document.getElementById('result-section');
    const resultMessage = document.getElementById('result-message');
    const resultPlayer = document.getElementById('result-player');
    const extractResult = document.getElementById('extract-result');
    const audioSource = document.getElementById('result-mp3-source');
    
    extractResult.classList.add('hidden');
    resultPlayer.classList.remove('hidden');
    
    currentStegoBlob = audioBlob;
    currentOriginalFilename = originalFilename;
    if (currentStegoUrl) {
        URL.revokeObjectURL(currentStegoUrl);
    }
    currentStegoUrl = URL.createObjectURL(audioBlob);
    
    resultMessage.innerHTML = `
        <div class="p-4 rounded-lg font-bold border-2 border-green-200 bg-green-50 text-green-800">
            ${message}
        </div>
    `;
    
    audioSource.src = currentStegoUrl;
    
    const audioElement = resultPlayer.querySelector('audio');
    audioElement.load();
    
    resultSection.classList.remove('hidden');
}

function showExtractResult(message, extractBlob, contentType) {
    const resultSection = document.getElementById('result-section');
    const resultMessage = document.getElementById('result-message');
    const resultPlayer = document.getElementById('result-player');
    const extractResult = document.getElementById('extract-result');
    const extractPreview = document.getElementById('extract-preview');
    
    resultPlayer.classList.add('hidden');
    extractResult.classList.remove('hidden');
    
    currentExtractBlob = extractBlob;
    currentExtractContentType = contentType;
    if (currentExtractUrl) {
        URL.revokeObjectURL(currentExtractUrl);
    }
    currentExtractUrl = URL.createObjectURL(extractBlob);
    
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

async function downloadStegoFile() {
    if (!currentStegoBlob) return;
    
    const defaultFilename = currentOriginalFilename ? 
        'stego_' + currentOriginalFilename : 
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
            await writable.write(currentStegoBlob);
            await writable.close();
            
            showResult('File saved successfully!', false);
            return;
        } catch (err) {
            if (err.name !== 'AbortError') {
                console.error('Error saving file:', err);
            }
            return;
        }
    }
    
    const a = document.createElement('a');
    a.href = currentStegoUrl;
    a.download = defaultFilename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
}

async function downloadExtractedFile() {
    if (!currentExtractBlob) return;
    
    let defaultFilename = 'extracted_secret.txt';
    if (currentExtractContentType) {
        if (currentExtractContentType === 'text/plain') {
            defaultFilename = 'extracted_secret.txt';
        } else if (currentExtractContentType === 'image/jpeg') {
            defaultFilename = 'extracted_secret.jpg';
        } else if (currentExtractContentType === 'image/png') {
            defaultFilename = 'extracted_secret.png';
        } else if (currentExtractContentType === 'application/pdf') {
            defaultFilename = 'extracted_secret.pdf';
        } else {
            const mimeToExt = {
                'image/gif': '.gif',
                'image/bmp': '.bmp',
                'audio/mpeg': '.mp3',
                'audio/wav': '.wav',
                'video/mp4': '.mp4',
                'application/zip': '.zip',
                'application/json': '.json'
            };
            const ext = mimeToExt[currentExtractContentType] || '.dat';
            defaultFilename = 'extracted_secret' + ext;
        }
    }
    
    if ('showSaveFilePicker' in window) {
        try {   
            const extension = defaultFilename.split('.').pop().toLowerCase();
            let acceptTypes = {};
            let description = 'All files';
            
            if (extension === 'txt') {
                acceptTypes = { 'text/plain': ['.txt'] };
                description = 'Text files';
            } else if (['jpg', 'jpeg'].includes(extension)) {
                acceptTypes = { 'image/jpeg': ['.jpg', '.jpeg'] };
                description = 'JPEG images';
            } else if (extension === 'png') {
                acceptTypes = { 'image/png': ['.png'] };
                description = 'PNG images';
            } else if (extension === 'pdf') {
                acceptTypes = { 'application/pdf': ['.pdf'] };
                description = 'PDF files';
            } else {
                acceptTypes = { '*/*': ['.' + extension] };
                description = 'All files';
            }
            
            const fileHandle = await window.showSaveFilePicker({
                suggestedName: defaultFilename,
                types: [{
                    description: description,
                    accept: acceptTypes
                }]
            });
            
            const writable = await fileHandle.createWritable();
            await writable.write(currentExtractBlob);
            await writable.close();
            
            showResult('File saved successfully!', false);
            return;
        } catch (err) {
            if (err.name !== 'AbortError') {
                console.error('Error saving file:', err);
            }
            return;
        }
    }
    
    const a = document.createElement('a');
    a.href = currentExtractUrl;
    a.download = defaultFilename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
}

async function handleEmbed(event) {
    event.preventDefault();
    
    const form = event.target;
    const keyInput = document.getElementById('embed-key');
    
    let hasInvalidChars = false;
    for (let i = 0; i < keyInput.value.length; i++) {
        if (keyInput.value.charCodeAt(i) > 255) {
            hasInvalidChars = true;
            break;
        }
    }
    
    if (hasInvalidChars) {
        showResult('Error: Key can only contain ASCII characters (0-255)', true);
        return;
    }
    
    if (keyInput.value.length < 1) {
        showResult('Error: Key cannot be empty', true);
        return;
    }
    
    const formData = new FormData(form);
    
    formData.set('use_encryption', document.getElementById('embed-encryption').checked);
    formData.set('use_key_for_position', document.getElementById('embed-position').checked);
    formData.set('lsb_bits', document.getElementById('embed-lsb').value);

    try {
        showResult('Processing... Please wait', false);
        
        const response = await fetch(`${API_BASE}/embed`, {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            const blob = await response.blob();
            const originalFilename = document.getElementById('embed-mp3-file').files[0].name;
            
            showEmbedResult('Secret file embedded successfully! Preview the result below:', blob, originalFilename);
        } else {
            const errorData = await response.json();
            throw new Error(errorData.message || 'Embedding failed');
        }
    } catch (error) {
        console.error('Error embedding file:', error);
        showResult('Error: ' + error.message, true);
    }
}

async function handleExtract(event) {
    event.preventDefault();
    
    const form = event.target;
    const keyInput = document.getElementById('extract-key');
    
    let hasInvalidChars = false;
    for (let i = 0; i < keyInput.value.length; i++) {
        if (keyInput.value.charCodeAt(i) > 255) {
            hasInvalidChars = true;
            break;
        }
    }
    
    if (hasInvalidChars) {
        showResult('Error: Key can only contain ASCII characters (0-255)', true);
        return;
    }
    
    if (keyInput.value.length < 1) {
        showResult('Error: Key cannot be empty', true);
        return;
    }
    
    const formData = new FormData(form);
    
    formData.set('use_encryption', document.getElementById('extract-encryption').checked);
    formData.set('use_key_for_position', document.getElementById('extract-position').checked);
    formData.set('lsb_bits', document.getElementById('extract-lsb').value);

    try {
        showResult('Extracting... Please wait', false);
        
        const response = await fetch(`${API_BASE}/extract`, {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            const blob = await response.blob();
            const contentType = response.headers.get('content-type') || 'application/octet-stream';
            
            showExtractResult('Secret file extracted successfully! Preview the result below:', blob, contentType);
        } else {
            const errorData = await response.json();
            throw new Error(errorData.message || 'Extraction failed');
        }
    } catch (error) {
        console.error('Error extracting file:', error);
        showResult('Error: ' + error.message, true);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    checkConnection();

    document.getElementById('embedForm').addEventListener('submit', handleEmbed);
    document.getElementById('extractForm').addEventListener('submit', handleExtract);

    document.getElementById('embed-key').addEventListener('input', function() {
        validateKey(this);
    });
    document.getElementById('extract-key').addEventListener('input', function() {
        validateKey(this);
    });

    document.getElementById('embed-mp3-file').addEventListener('change', function() {
        showFileInfo(this, 'embed-mp3-info', 'embed-mp3-player', 'embed-mp3-source');
    });
    document.getElementById('embed-secret-file').addEventListener('change', function() {
        showFileInfo(this, 'embed-secret-info');
    });
    document.getElementById('extract-mp3-file').addEventListener('change', function() {
        showFileInfo(this, 'extract-mp3-info', 'extract-mp3-player', 'extract-mp3-source');
    });
});