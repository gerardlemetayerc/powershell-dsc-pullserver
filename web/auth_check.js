
// Vérifie la présence d'un token JWT dans localStorage, sinon redirige vers /web/login
(function() {
    if (!localStorage.getItem('jwt_token')) {
        window.location.href = '/web/login';
    }
})();

