// web/user_session.js
// Gère l'affichage du menu utilisateur, la récupération du rôle, la déconnexion et l'activation du menu selon l'URL

function clearLegacyBrowserAuthState() {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('jwt_exp');
    document.cookie = 'jwt_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
    document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
}

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
                    clearLegacyBrowserAuthState();
                    window.location.href = '/web/login';
                });
        };
    }
});

document.addEventListener('DOMContentLoaded', function() {
    const aboutMenuItem = document.getElementById('aboutMenuItem');
    if (!aboutMenuItem) {
        return;
    }

    aboutMenuItem.addEventListener('click', function(event) {
        event.preventDefault();

        fetch('/api/v1/about', { credentials: 'same-origin' })
            .then(function(resp) {
                if (!resp.ok) {
                    throw new Error('about fetch failed');
                }
                return resp.json();
            })
            .then(function(info) {
                document.getElementById('aboutBuildVersion').textContent = info.build_version || '-';
                document.getElementById('aboutDbVersion').textContent = info.db_version || '-';
                $('#aboutModal').modal('show');
            })
            .catch(function() {
                document.getElementById('aboutBuildVersion').textContent = 'Unavailable';
                document.getElementById('aboutDbVersion').textContent = 'Unavailable';
                $('#aboutModal').modal('show');
            });
    });
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
