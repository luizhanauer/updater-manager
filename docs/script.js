// Gera elementos flutuantes aleatórios para o fundo
const createFloatingPackages = () => {
    const container = document.getElementById('background-animation');
    const packageCount = 20; // Quantidade de pacotes na tela
    const icons = ['fa-box', 'fa-box-open', 'fa-cube', 'fa-cubes'];

    for (let i = 0; i < packageCount; i++) {
        const el = document.createElement('i');
        const randomIcon = icons[Math.floor(Math.random() * icons.length)];
        
        el.classList.add('fas', randomIcon, 'floating-package');
        
        // Randomiza tamanho, posição horizontal, duração da animação e atraso inicial
        const size = Math.random() * 2 + 1; // Entre 1rem e 3rem
        const left = Math.random() * 100; // Entre 0% e 100% da largura
        const duration = Math.random() * 15 + 10; // Entre 10s e 25s
        const delay = Math.random() * 10; // Entre 0s e 10s

        el.style.fontSize = `${size}rem`;
        el.style.left = `${left}%`;
        el.style.animationDuration = `${duration}s`;
        el.style.animationDelay = `${delay}s`;

        container.appendChild(el);
    }
};

// Lida com a cópia para a área de transferência
const copyCommand = async () => {
    const commandText = document.getElementById('installCommand').innerText;
    const btn = document.getElementById('copyBtn');
    const icon = btn.querySelector('i');

    try {
        await navigator.clipboard.writeText(commandText);
        
        // Feedback visual de sucesso
        icon.className = 'fas fa-check';
        btn.classList.add('success');

        setTimeout(() => {
            icon.className = 'far fa-copy';
            btn.classList.remove('success');
        }, 2000);
        
    } catch (err) {
        console.error('Falha ao copiar comando: ', err);
        
        // Feedback visual de erro
        icon.className = 'fas fa-times';
        icon.style.color = '#ef4444';
        
        setTimeout(() => {
            icon.className = 'far fa-copy';
            icon.style.color = '';
        }, 2000);
    }
};

// Inicializa o fundo animado quando o DOM estiver pronto
document.addEventListener('DOMContentLoaded', () => {
    createFloatingPackages();
});