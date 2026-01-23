// Si le token JWT est invalide (401 sur une requête AJAX), on efface la session et on redirige vers /web/login
$(document).ajaxError(function(event, jqxhr, settings, thrownError) {
    if (jqxhr.status === 401) {
        localStorage.removeItem('jwt_token');
        window.location.href = '/web/login';
    }
});
