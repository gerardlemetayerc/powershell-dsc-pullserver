// Si le token JWT est invalide (401 sur une requête AJAX), on redirige vers /web/login
$(document).ajaxError(function(event, jqxhr, settings, thrownError) {
    if (jqxhr.status === 401) {
        window.location.href = '/web/login';
    }
});
