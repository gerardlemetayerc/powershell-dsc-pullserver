// web/user_session.js
// Gère l'affichage du menu utilisateur, la récupération du rôle, la déconnexion et l'activation du menu selon l'URL

// Affiche le nom d'utilisateur et le menu admin selon l'API /api/v1/my (JWT via cookie)
document.addEventListener('DOMContentLoaded', function() {
    fetch('/api/v1/my', { credentials: 'same-origin' })
        .then(function(resp) {
            if (!resp.ok) throw new Error('not auth');
            return resp.json();
        })
        .then(function(user) {
            // Stocke l'utilisateur globalement
            window.currentUser = user;
            // Affiche l'email dans le menu
            if (user && user.email) {
                document.getElementById('userName').textContent = user.email;
            }
            // Affiche ou masque le menu admin selon le rôle
            var adminMenu = document.getElementById('adminMenu');
            if (adminMenu) {
                if (user && user.role === 'admin') {
                    adminMenu.style.display = '';
                } else {
                    adminMenu.style.display = 'none';
                }
            }
            // Signale que l'utilisateur est prêt
            window.dispatchEvent(new Event('userReady'));
        })
        .catch(function() {
            // Si non authentifié, redirige vers login
            window.location.href = '/web/login';
        });
});
// Déconnexion
document.addEventListener('DOMContentLoaded', function() {
    const btn = document.getElementById('logoutBtn');
    if (btn) {
        btn.onclick = function() {
            // Appelle le backend pour supprimer le cookie JWT (HttpOnly)
            fetch('/api/v1/logout', { method: 'POST', credentials: 'same-origin' })
                .finally(function() {
                    window.location.href = '/web/login';
                });
        };
    }
});
// A propos
document.addEventListener('DOMContentLoaded', function() {
    const aboutBtn = document.getElementById('aboutBtn');
    if (!aboutBtn) return;
    aboutBtn.onclick = function(e) {
        e.preventDefault();
        fetch('/api/v1/version', { credentials: 'same-origin' })
            .then(function(resp) {
                if (!resp.ok) throw new Error('version unavailable');
                return resp.json();
            })
            .then(function(payload) {
                document.getElementById('about-app-version').textContent = payload.version || '-';
                document.getElementById('about-db-version').textContent = payload.target_db_version || '-';
                $('#aboutModal').modal('show');
            })
            .catch(function() {
                document.getElementById('about-app-version').textContent = 'N/A';
                document.getElementById('about-db-version').textContent = 'N/A';
                $('#aboutModal').modal('show');
            });
    };
});
// Active dynamiquement le menu selon l'URL
document.addEventListener('DOMContentLoaded', function() {
    var path = window.location.pathname;
    document.querySelectorAll('.main-sidebar .nav-link').forEach(function(link) {
        // Retire la classe active de tous
        link.classList.remove('active');
        // Active si correspond à l'URL courante
        if (link.getAttribute('href') === path) {
            link.classList.add('active');
        }
    });
});
