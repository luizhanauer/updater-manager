function copyCommand() {
    const commandText = document.getElementById('cmd').innerText;
    
    navigator.clipboard.writeText(commandText).then(() => {
        // Feedback Visual
        const toast = document.getElementById('toast');
        const btnIcon = document.querySelector('#copyBtn i');
        
        // Mostra Toast
        toast.classList.add('show');
        
        // Troca ícone temporariamente
        // Nota: feather icons precisam ser rerenderizados ou manipulados via class
        // Aqui vamos simplificar mantendo o toast como feedback principal
        
        setTimeout(() => {
            toast.classList.remove('show');
        }, 2000);
    }).catch(err => {
        console.error('Falha ao copiar:', err);
    });
}