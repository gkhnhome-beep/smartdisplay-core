// Yumuşak, düşük frekanslı, melodik olmayan presence sesi için Web Audio API ile preload edilen buffer
// Bu dosya sadece presence sound kodunu içerir, ana dosyada import edilmeden doğrudan kullanılacak.

export function playPresenceSound() {
    if (window.__presenceSoundPlayed) return;
    window.__presenceSoundPlayed = true;
    if (typeof alarmLastState === 'object' && (alarmLastState.triggered || alarmLastState.state === 'triggered')) return;
    try {
        const ctx = new (window.AudioContext || window.webkitAudioContext)();
        const duration = 0.45; // 450ms
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.value = 220; // Düşük frekans (A3)
        gain.gain.setValueAtTime(0.0001, ctx.currentTime);
        gain.gain.linearRampToValueAtTime(0.18, ctx.currentTime + 0.04); // Yumuşak giriş
        gain.gain.linearRampToValueAtTime(0.09, ctx.currentTime + duration * 0.7); // Pulse
        gain.gain.linearRampToValueAtTime(0.0001, ctx.currentTime + duration); // Yumuşak çıkış
        osc.connect(gain).connect(ctx.destination);
        osc.start();
        osc.stop(ctx.currentTime + duration);
        osc.onended = () => ctx.close();
    } catch (e) {}
}
