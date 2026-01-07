import { Uppy, Dashboard, Tus } from "https://releases.transloadit.com/uppy/v3.0.0/uppy.min.mjs"

// saved token
const savedSecret = localStorage.getItem('odv_secret');
if(savedSecret) {
    const authBox = document.getElementById('auth-box');
    if (authBox) authBox.classList.add('hidden');
}

// save button logic
const saveBtn = document.getElementById('save-btn');
if (saveBtn) {
    saveBtn.addEventListener('click', () => {
        const val = document.getElementById('secret-input').value;
        localStorage.setItem('odv_secret', val);
        location.reload();
    });
}

// uppy config
const uppy = new Uppy()
    .use(Dashboard, { 
        inline: true, 
        target: '#uploader', 
        theme: 'dark',
        height: 400 
    })
    .use(Tus, { 
        endpoint: '/files/', 
        chunkSize: 5 * 1024 * 1024, // 5MB for cloudflare
        headers: {
            'X-Auth-Token': savedSecret || ''
        }
    })

uppy.on('complete', (result) => {
    const resultsDiv = document.getElementById('results');
    result.successful.forEach(file => {
        const uploadURL = file.uploadURL;
        const fileName = file.name;
        
        const box = document.createElement('div');
        box.className = 'link-box';

        const pName = document.createElement('p');
        pName.textContent = 'Fisier: ' + fileName;
        pName.style.fontWeight = 'bold';

        const pLink = document.createElement('p');
        const link = document.createElement('a');
        link.href = uploadURL;
        link.target = '_blank';
        link.textContent = uploadURL;
        pLink.textContent = 'Link: ';
        pLink.appendChild(link);

        const small = document.createElement('small');
        small.textContent = 'Expira in 24h';

        box.appendChild(pName);
        box.appendChild(pLink);
        box.appendChild(small);
        resultsDiv.appendChild(box);
    });
})