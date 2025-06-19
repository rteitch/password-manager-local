// Variabel global dan ikon SVG tetap sama
let allAccounts = [];
const eyeIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>`;
const eyeOffIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/><path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/><path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/><line x1="2" x2="22" y1="2" y2="22"/></svg>`;

document.addEventListener('DOMContentLoaded', async () => {
    try {
        const response = await fetch('/session-check');
        if (response.ok) {
            showMainApp();
            await loadAccounts();
        } else {
            showAuthSection('login-section');
        }
    } catch (error) {
        console.error("Gagal memeriksa sesi:", error);
        showAuthSection('login-section');
    }
});

// PERBAIKAN: Fungsi displayAccounts sekarang menambahkan data-label
function displayAccounts(accountsToDisplay) {
    const accountsListContainer = document.getElementById('accounts-list');
    accountsListContainer.innerHTML = ''; 

    if (!accountsToDisplay || accountsToDisplay.length === 0) {
        const p = document.createElement('p');
        p.style.textAlign = 'center';
        p.textContent = 'Tidak ada akun yang cocok atau belum ada akun yang disimpan.';
        accountsListContainer.appendChild(p);
        return;
    }

    const tableContainer = document.createElement('div');
    tableContainer.className = 'table-container';
    const table = document.createElement('table');
    table.className = 'account-table';

    const thead = table.createTHead();
    const headerRow = thead.insertRow();
    const headers = ['#', 'Layanan', 'Kategori', 'Username', 'Password', 'Tindakan'];
    headers.forEach(headerText => {
        const th = document.createElement('th');
        th.textContent = headerText;
        headerRow.appendChild(th);
    });

    const tbody = table.createTBody();
    accountsToDisplay.forEach((account, displayIndex) => {
        const originalIndex = allAccounts.findIndex(acc => acc.createdAt === account.createdAt && acc.service === account.service);
        const row = tbody.insertRow();
        
        // Menambahkan atribut data-label ke setiap sel
        row.insertCell().textContent = displayIndex + 1;
        
        const serviceCell = row.insertCell();
        serviceCell.textContent = account.service;
        serviceCell.setAttribute('data-label', 'Layanan');

        const categoryCell = row.insertCell();
        categoryCell.textContent = account.category || '-';
        categoryCell.setAttribute('data-label', 'Kategori');

        const usernameCell = row.insertCell();
        usernameCell.textContent = account.username;
        usernameCell.setAttribute('data-label', 'Username');

        const passwordCell = row.insertCell();
        passwordCell.setAttribute('data-label', 'Password');
        const passwordSpan = document.createElement('span');
        passwordSpan.className = 'password-text';
        passwordSpan.textContent = '••••••••';
        const showHideBtn = document.createElement('button');
        showHideBtn.className = 'btn-show-hide';
        showHideBtn.innerHTML = eyeIconSVG;
        showHideBtn.onclick = () => {
            if (passwordSpan.textContent === '••••••••') {
                passwordSpan.textContent = account.password;
                showHideBtn.innerHTML = eyeOffIconSVG;
            } else {
                passwordSpan.textContent = '••••••••';
                showHideBtn.innerHTML = eyeIconSVG;
            }
        };
        passwordCell.append(passwordSpan, showHideBtn);

        const actionCell = row.insertCell();
        actionCell.setAttribute('data-label', 'Tindakan');
        const actionContainer = document.createElement('div');
        actionContainer.className = 'action-buttons';
        const copyButton = document.createElement('button');
        copyButton.className = 'btn-copy';
        copyButton.textContent = 'Salin';
        copyButton.onclick = () => copyToClipboard(account.password, account.service);
        const editButton = document.createElement('button');
        editButton.className = 'btn-edit';
        editButton.textContent = 'Ubah';
        editButton.onclick = () => openEditModal(originalIndex);
        const deleteButton = document.createElement('button');
        deleteButton.className = 'btn-delete';
        deleteButton.textContent = 'Hapus';
        deleteButton.onclick = () => deleteAccount(originalIndex, account.service);
        actionContainer.append(copyButton, editButton, deleteButton);
        actionCell.appendChild(actionContainer);
    });

    tableContainer.appendChild(table);
    accountsListContainer.appendChild(tableContainer);
}

// Fungsi lain tidak berubah...
function showAuthSection(sectionId) { document.getElementById('auth-container').style.display = 'block'; document.getElementById('main-section').style.display = 'none'; document.querySelectorAll('.auth-section').forEach(section => section.style.display = 'none'); document.getElementById(sectionId).style.display = 'block'; }
function showMainApp() { document.getElementById('auth-container').style.display = 'none'; document.getElementById('main-section').style.display = 'block'; showSection('list-accounts-section', document.getElementById('nav-list')); }
function showSection(sectionId, clickedButton) { document.querySelectorAll('.app-content').forEach(section => section.style.display = 'none'); document.getElementById(sectionId).style.display = 'block'; if (clickedButton) { document.querySelectorAll('#app-nav button').forEach(button => button.classList.remove('active')); clickedButton.classList.add('active'); } if (sectionId === 'list-accounts-section') { document.getElementById('searchInput').value = ''; filterAccounts(); } }
function openEditModal(index) { const account = allAccounts[index]; if (!account) { Swal.fire('Error', 'Akun tidak ditemukan.', 'error'); return; } Swal.fire({ title: `Ubah Akun`, html: `<div class="form-group"><input id="swal-service" class="swal2-input" value="${account.service}" placeholder="Nama Layanan"></div>` + `<div class="form-group"><input id="swal-username" class="swal2-input" value="${account.username}" placeholder="Username / Email"></div>` + `<div class="form-group"><input id="swal-password" class="swal2-input" value="${account.password}" placeholder="Password"></div>` + `<div class="form-group"><input id="swal-category" class="swal2-input" value="${account.category || ''}" placeholder="Kategori"></div>` + `<div class="form-group"><textarea id="swal-notes" class="swal2-textarea" placeholder="Catatan">${account.notes || ''}</textarea></div>`, focusConfirm: false, showCancelButton: true, confirmButtonText: 'Simpan Perubahan', cancelButtonText: 'Batal', customClass: { popup: 'kali-swal' }, preConfirm: () => { return { index, service: document.getElementById('swal-service').value, username: document.getElementById('swal-username').value, password: document.getElementById('swal-password').value, category: document.getElementById('swal-category').value, notes: document.getElementById('swal-notes').value } }, }).then(async (result) => { if (result.isConfirmed) { const newData = result.value; if (!newData.service || !newData.username || !newData.password) { Swal.fire('Error', 'Layanan, Username, dan Password tidak boleh kosong.', 'error'); return; } try { const response = await fetch('/update', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(newData) }); if (response.ok) { Swal.fire('Berhasil!', 'Akun telah diperbarui.', 'success'); await loadAccounts(); } else { const error = await response.text(); Swal.fire('Gagal!', `Gagal memperbarui akun: ${error}`, 'error'); } } catch (error) { Swal.fire('Gagal!', 'Terjadi kesalahan jaringan.', 'error'); } } }); }
function copyToClipboard(textToCopy, serviceName) { navigator.clipboard.writeText(textToCopy).then(() => { Swal.fire({ icon: 'success', title: 'Disalin!', text: `Password untuk ${serviceName} telah disalin.`, toast: true, position: 'top-end', showConfirmButton: false, timer: 2000, timerProgressBar: true }); }).catch(err => { console.error('Gagal menyalin teks: ', err); }); }
async function loadAccounts() { try { const response = await fetch('/list', { method: 'POST' }); if (response.ok) { allAccounts = await response.json() || []; allAccounts.sort((a, b) => a.service.localeCompare(b.service)); displayAccounts(allAccounts); } else if (response.status === 401) { location.reload(); } else { document.getElementById('accounts-list').innerHTML = '<p style="text-align:center; color: red;">Gagal memuat akun.</p>'; } } catch (error) { console.error("Gagal memuat akun:", error); } }
function filterAccounts() { const searchTerm = document.getElementById('searchInput').value.toLowerCase(); const filtered = searchTerm ? allAccounts.filter(acc => acc.service.toLowerCase().includes(searchTerm) || acc.username.toLowerCase().includes(searchTerm) || (acc.category && acc.category.toLowerCase().includes(searchTerm))) : allAccounts; displayAccounts(filtered); }
async function deleteAccount(index, serviceName) { Swal.fire({ title: `Hapus Akun ${serviceName}?`, text: "Tindakan ini tidak dapat dibatalkan!", icon: 'warning', showCancelButton: true, confirmButtonColor: '#d33', cancelButtonColor: '#3085d6', confirmButtonText: 'Ya, Hapus!', cancelButtonText: 'Batal' }).then(async (result) => { if (result.isConfirmed) { const response = await fetch('/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ index: index }) }); if (response.ok) { Swal.fire('Dihapus!', `Akun ${serviceName} telah dihapus.`, 'success'); await loadAccounts(); } else if (response.status === 401) { Swal.fire({ icon: 'warning', title: 'Sesi Berakhir', text: 'Silakan login kembali.' }).then(() => location.reload()); } else { Swal.fire({ icon: 'error', title: 'Gagal', text: 'Gagal menghapus akun.' }); } } }); }
async function resetPassword(event) { event.preventDefault(); const oldPassword = document.getElementById('oldPassword').value, newPassword = document.getElementById('newPassword').value, confirmPassword = document.getElementById('confirmPassword').value; if (newPassword !== confirmPassword) { Swal.fire({ icon: 'error', title: 'Tidak Cocok', text: 'Konfirmasi password baru tidak cocok!' }); return; } const response = await fetch('/reset-password', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ oldPassword, newPassword }) }); if (response.ok) { Swal.fire({ icon: 'success', title: 'Berhasil Direset!', text: 'Password Anda telah diperbarui. Anda akan logout.', }).then(() => logout()); event.target.reset(); } else if (response.status === 401) { Swal.fire({ icon: 'warning', title: 'Sesi Berakhir', text: 'Silakan login kembali.' }).then(() => location.reload()); } else { const error = await response.text(); Swal.fire({ icon: 'error', title: 'Error', text: error }); } }
function updateLengthValue(value) { document.getElementById('length-value').innerText = value; }
function generatePassword() { const length = parseInt(document.getElementById('length').value), includeUppercase = document.getElementById('uppercase').checked, includeLowercase = document.getElementById('lowercase').checked, includeNumbers = document.getElementById('numbers').checked, includeSymbols = document.getElementById('symbols').checked; const upperChars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', lowerChars = 'abcdefghijklmnopqrstuvwxyz', numberChars = '0123456789', symbolChars = '!@#$%^&*()_+-=[]{}|;:,.<>?'; let charPool = ''; if (includeUppercase) charPool += upperChars; if (includeLowercase) charPool += lowerChars; if (includeNumbers) charPool += numberChars; if (includeSymbols) charPool += symbolChars; if (charPool === '') { Swal.fire({ icon: 'error', title: 'Tidak Ada Karakter', text: 'Pilih setidaknya satu jenis karakter.' }); return; } let password = ''; const secureRandomValues = new Uint32Array(length); window.crypto.getRandomValues(secureRandomValues); for (let i = 0; i < length; i++) { password += charPool[secureRandomValues[i] % charPool.length]; } document.getElementById('generated-password-output').value = password; }
async function addAccount(event) { event.preventDefault(); const service = document.getElementById('service').value, username = document.getElementById('username').value, password = document.getElementById('password').value, category = document.getElementById('category').value, notes = document.getElementById('notes').value; const response = await fetch('/add', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ service, username, password, category, notes }) }); if (response.ok) { Swal.fire({ icon: 'success', title: 'Berhasil!', text: 'Akun baru telah ditambahkan.', timer: 1500, showConfirmButton: false }); event.target.reset(); await loadAccounts(); showSection('list-accounts-section', document.getElementById('nav-list')); } else if (response.status === 401) { Swal.fire({ icon: 'warning', title: 'Sesi Berakhir', text: 'Silakan login kembali.' }).then(() => location.reload()); } else { Swal.fire({ icon: 'error', title: 'Oops...', text: 'Gagal menambahkan akun!' }); } }
async function register(event) { event.preventDefault(); const username = document.getElementById('regUsername').value; const email = document.getElementById('regEmail').value; const password = document.getElementById('regPassword').value; try { const response = await fetch('/register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username, email, password }), }); const message = await response.text(); if (response.ok) { Swal.fire({ icon: 'success', title: 'Registrasi Berhasil', text: message }); event.target.reset(); showAuthSection('login-section'); } else { Swal.fire({ icon: 'error', title: 'Gagal Registrasi', text: message }); } } catch (error) { Swal.fire({ icon: 'error', title: 'Oops...', text: 'Terjadi kesalahan jaringan.' }); } }
async function login(event) { event.preventDefault(); const username = document.getElementById('loginUsername').value; const password = document.getElementById('loginPassword').value; try { const response = await fetch('/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username, password }), }); if (response.ok) { event.target.reset(); showMainApp(); await loadAccounts(); } else { const message = await response.text(); Swal.fire({ icon: 'error', title: 'Gagal Login', text: message }); } } catch (error) { Swal.fire({ icon: 'error', title: 'Oops...', text: 'Terjadi kesalahan jaringan.' }); } }
async function logout() { Swal.fire({ title: 'Anda Yakin?', text: "Anda akan keluar dari sesi ini.", icon: 'warning', showCancelButton: true, confirmButtonColor: '#3085d6', cancelButtonColor: '#d33', confirmButtonText: 'Ya, Logout!', cancelButtonText: 'Batal' }).then(async (result) => { if (result.isConfirmed) { await fetch('/logout', { method: 'POST' }); location.reload(); } }); }
