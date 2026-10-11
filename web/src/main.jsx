import React, { useEffect, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { dataUrlToFile, postStage } from './api.js';
import ShaderBg from './ShaderBg.jsx';
import './index.css';

// Custom SVG Icons (Zero emojis used anywhere)
function SpiderIcon({ className = 'w-6 h-6' }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 2c-.6 0-1.1.4-1.3 1L9.2 6.5C7.5 7.1 6.1 8.3 5.3 9.9L2.4 8.7c-.5-.2-1 .1-1.2.6-.2.5.1 1 .6 1.2l3 1.2c-.1.5-.1 1-.1 1.5 0 .7.1 1.4.3 2l-3.2 1.9c-.4.3-.6.8-.3 1.3.2.4.6.6 1 .6.2 0 .4 0 .6-.1l3-1.8c1.1 1.7 2.8 3 4.9 3.4l-.6 2.6c-.1.5.2 1 .7 1.1h.2c.5 0 .9-.3 1-.8l.7-3h1.4l.7 3c.1.5.5.8 1 .8h.2c.5-.1.8-.6.7-1.1l-.6-2.6c2.1-.4 3.8-1.7 4.9-3.4l3 1.8c.2.1.4.1.6.1.4 0 .8-.2 1-.6.3-.5.1-1-.3-1.3l-3.2-1.9c.2-.6.3-1.3.3-2 0-.5 0-1-.1-1.5l3-1.2c.5-.2.8-.7.6-1.2-.2-.5-.7-.8-1.2-.6l-2.9 1.2c-.8-1.6-2.2-2.8-3.9-3.4L13.3 3c-.2-.6-.7-1-1.3-1zm0 7c1.7 0 3 1.3 3 3 0 .8-.3 1.5-.8 2-.5.5-1.2.8-2.2.8s-1.7-.3-2.2-.8c-.5-.5-.8-1.2-.8-2 0-1.7 1.3-3 3-3zm0 2c-.6 0-1 .4-1 1s.4 1 1 1 1-.4 1-1-.4-1-1-1z" />
    </svg>
  );
}

function UploadIcon({ className = 'w-5 h-5' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
    </svg>
  );
}

function DownloadIcon({ className = 'w-4 h-4' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
    </svg>
  );
}

function TargetIcon({ className = 'w-5 h-5' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="9" strokeWidth="2" />
      <circle cx="12" cy="12" r="4" strokeWidth="2" />
      <path strokeLinecap="round" strokeWidth="2" d="M12 2v3m0 14v3M2 12h3m14 0h3" />
    </svg>
  );
}

function LockIcon({ className = 'w-4 h-4' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <rect x="5" y="11" width="14" height="10" rx="2" strokeWidth="2" />
      <path strokeLinecap="round" strokeWidth="2" d="M8 11V7a4 4 0 018 0v4" />
    </svg>
  );
}

function EyeIcon({ className = 'w-4 h-4' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
    </svg>
  );
}

function EyeOffIcon({ className = 'w-4 h-4' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l18 18" />
    </svg>
  );
}

function CrosshairBoltIcon({ className = 'w-5 h-5' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
    </svg>
  );
}

function AlertIcon({ className = 'w-5 h-5' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>
  );
}

function RadarGridIcon({ className = 'w-5 h-5' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <rect x="3" y="3" width="18" height="18" rx="2" strokeWidth="2" />
      <path strokeWidth="1.5" d="M3 9h18M3 15h18M9 3v18M15 3v18" />
    </svg>
  );
}

function ZoomIcon({ className = 'w-4 h-4' }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="11" cy="11" r="7" strokeWidth="2" />
      <path strokeLinecap="round" strokeWidth="2.5" d="M20 20l-4-4m-5-1a3 3 0 100-6 3 3 0 000 6z" />
    </svg>
  );
}

const initialImages = {
  original: null,
  watermarked: null,
  tampered: null,
  overlay: null,
  recovered: null,
};

// Comic metric panel component
function ComicMetric({ label, value, unit, status }) {
  return (
    <div className="relative overflow-hidden rounded-lg border-2 border-comic-border bg-comic-card p-3 shadow-comic-sm transition-all hover:border-comic-blue">
      <div className="absolute right-0 top-0 h-6 w-6 translate-x-3 -translate-y-3 rotate-45 bg-comic-red/40" />
      <div className="font-display text-[10px] font-bold uppercase tracking-widest text-comic-muted">
        {label}
      </div>
      <div className="mt-1 flex items-baseline gap-1 font-mono text-xl font-black tabular-nums text-comic-paper">
        {typeof value === 'number' && Number.isFinite(value)
          ? value.toFixed(unit === '%' ? 1 : 3)
          : value ?? '—'}
        {value != null && unit && (
          <span className="font-display text-xs font-semibold text-comic-blue">{unit}</span>
        )}
      </div>
      {status && (
        <div className="mt-1 font-mono text-[9px] uppercase tracking-wider text-comic-yellow">
          [{status}]
        </div>
      )}
    </div>
  );
}

// Comic Stage Image Panel
function ComicStageImage({ step, label, src, hint, onZoom }) {
  return (
    <section className="group relative flex flex-col overflow-hidden rounded-xl border-2 border-comic-border bg-comic-card shadow-comic transition-all hover:border-comic-red hover:shadow-comic-red">
      {/* Comic Header Bar */}
      <div className="flex items-center justify-between border-b-2 border-comic-border bg-comic-panel px-3.5 py-2.5">
        <div className="flex items-center gap-2 min-w-0">
          <span className="shrink-0 rounded bg-comic-red px-1.5 py-0.5 font-comic text-xs tracking-wider text-white">
            {step}
          </span>
          <h4 className="truncate font-display text-xs font-bold uppercase tracking-wider text-comic-paper">
            {label}
          </h4>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {src && (
            <>
              <button
                type="button"
                onClick={() => onZoom(src, label)}
                title="Expand image"
                className="rounded border border-comic-border bg-comic-panelAlt p-1 text-comic-muted hover:border-comic-blue hover:text-comic-blue"
              >
                <ZoomIcon />
              </button>
              <a
                href={src}
                download={`${label.toLowerCase().replaceAll(' ', '-')}.png`}
                title="Download PNG"
                className="flex items-center gap-1 rounded border border-comic-red/40 bg-comic-red/10 px-2 py-0.5 font-display text-[10px] font-bold uppercase text-comic-red hover:bg-comic-red hover:text-white"
              >
                <DownloadIcon />
                <span className="hidden sm:inline">PNG</span>
              </a>
            </>
          )}
        </div>
      </div>

      {/* Image / Canvas Display */}
      <div className="relative flex aspect-[4/3] w-full items-center justify-center overflow-hidden bg-comic-ink/80 p-2 bg-halftone-dots">
        {src ? (
          <img
            src={src}
            alt={label}
            className="max-h-full max-w-full rounded object-contain shadow-sm cursor-pointer transition-transform duration-200 group-hover:scale-[1.01]"
            onClick={() => onZoom(src, label)}
          />
        ) : (
          <div className="flex flex-col items-center justify-center p-4 text-center">
            <div className="mb-2 rounded-full border border-dashed border-comic-border p-3 text-comic-muted">
              <SpiderIcon className="w-6 h-6 opacity-40" />
            </div>
            <p className="max-w-[200px] font-sans text-xs text-comic-muted">{hint}</p>
          </div>
        )}
      </div>
    </section>
  );
}

// Tamper Grid Visualizer (Canvas renderer)
function TamperGrid({ map, report }) {
  const canvasRef = useRef(null);
  const blocksX = report?.blocksX ?? map?.[0]?.length ?? 0;
  const blocksY = report?.blocksY ?? map?.length ?? 0;
  const blockSize = report?.blockSize || 8;
  const width = report?.imgWidth || blocksX * blockSize;
  const height = report?.imgHeight || blocksY * blockSize;

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !map?.length || !width || !height) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, width, height);

    map.forEach((row, y) => {
      row.forEach((flag, x) => {
        const px = x * blockSize;
        const py = y * blockSize;
        const tileWidth = Math.min(blockSize, width - px);
        const tileHeight = Math.min(blockSize, height - py);
        if (tileWidth <= 0 || tileHeight <= 0) return;

        if (flag) {
          // Flagged as Tampered - Vivid Comic Crimson
          ctx.fillStyle = '#ff2453';
          ctx.fillRect(px, py, tileWidth, tileHeight);
          ctx.strokeStyle = 'rgba(255, 255, 255, 0.4)';
          ctx.lineWidth = Math.max(0.5, blockSize / 12);
          ctx.strokeRect(px, py, tileWidth, tileHeight);
        } else {
          // Verified Intact - Deep Cyan-Navy Comic Grid
          ctx.fillStyle = 'rgba(16, 26, 48, 0.92)';
          ctx.fillRect(px, py, tileWidth, tileHeight);
          ctx.strokeStyle = 'rgba(0, 229, 255, 0.15)';
          ctx.lineWidth = Math.max(0.25, blockSize / 20);
          ctx.strokeRect(px, py, tileWidth, tileHeight);
        }
      });
    });
  }, [map, width, height, blockSize]);

  if (!map?.length) return null;

  const flaggedCount = report?.flaggedBlocks ?? 0;
  const totalCount = report?.totalBlocks ?? blocksX * blocksY;
  const tamperRatio = totalCount > 0 ? ((flaggedCount / totalCount) * 100).toFixed(1) : 0;

  return (
    <section className="relative overflow-hidden rounded-xl border-2 border-comic-border bg-comic-card p-4 shadow-comic">
      {/* Corner badge */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b-2 border-comic-border pb-3">
        <div className="flex items-center gap-2">
          <RadarGridIcon className="text-comic-red" />
          <h3 className="font-display text-sm font-bold uppercase tracking-wider text-comic-paper">
            Tamper Radar Matrix
          </h3>
          <span className="rounded bg-comic-red/20 px-2 py-0.5 font-mono text-[10px] font-bold text-comic-red border border-comic-red/40">
            {flaggedCount} FLAGGED ({tamperRatio}%)
          </span>
        </div>
        <div className="flex items-center gap-3 font-mono text-xs text-comic-muted">
          <span>{blocksX} × {blocksY} BLOCKS</span>
          <span className="hidden sm:inline">|</span>
          <span className="hidden sm:inline">{blockSize} × {blockSize} PX / TILE</span>
        </div>
      </div>

      <div className="mt-4 flex justify-center overflow-x-auto rounded-lg border-2 border-comic-border bg-comic-ink p-2">
        <canvas
          ref={canvasRef}
          width={width}
          height={height}
          className="max-h-[380px] w-auto max-w-full rounded object-contain shadow-inner"
          role="img"
          aria-label="Tamper radar: red cells represent flagged tampered blocks"
        />
      </div>

      <div className="mt-3 flex flex-wrap items-center justify-between gap-2 text-xs">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1.5">
            <span className="h-3 w-3 rounded-sm bg-comic-red border border-white/50" />
            <span className="font-display font-medium text-comic-paper">Compromised Block</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="h-3 w-3 rounded-sm bg-[#101a30] border border-comic-blue/40" />
            <span className="font-display font-medium text-comic-muted">Verified Intact</span>
          </div>
        </div>
        <div className="font-mono text-[10px] text-comic-muted">
          CANVAS RESOLUTION: {width} × {height} PX
        </div>
      </div>
    </section>
  );
}

// Image Zoom Modal
function ImageModal({ image, label, onClose }) {
  if (!image) return null;
  return (
    <div
      role="dialog"
      aria-modal="true"
      onClick={onClose}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4 backdrop-blur-sm"
    >
      <div
        onClick={e => e.stopPropagation()}
        className="relative max-h-[92vh] max-w-[94vw] overflow-hidden rounded-xl border-2 border-comic-red bg-comic-panel p-3 shadow-comic-lg"
      >
        <div className="mb-2 flex items-center justify-between border-b border-comic-border pb-2">
          <div className="font-display text-sm font-bold uppercase text-comic-paper">{label}</div>
          <div className="flex items-center gap-2">
            <a
              href={image}
              download={`${label.toLowerCase().replaceAll(' ', '-')}.png`}
              className="rounded bg-comic-red px-2.5 py-1 font-display text-xs font-bold uppercase text-white hover:bg-comic-redDark"
            >
              Download PNG
            </a>
            <button
              onClick={onClose}
              className="rounded border border-comic-border px-2 py-1 font-mono text-xs font-bold text-comic-muted hover:border-comic-red hover:text-white"
            >
              ESC / Close
            </button>
          </div>
        </div>
        <div className="flex items-center justify-center overflow-auto max-h-[82vh]">
          <img src={image} alt={label} className="max-h-full max-w-full rounded object-contain" />
        </div>
      </div>
    </div>
  );
}

function App() {
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [key, setKey] = useState('stegno');
  const [showKey, setShowKey] = useState(false);
  const [region, setRegion] = useState({ x: '8', y: '8', w: '32', h: '32', mode: 'black' });
  const [images, setImages] = useState(initialImages);
  const [embedMetrics, setEmbedMetrics] = useState(null);
  const [verifyReport, setVerifyReport] = useState(null);
  const [recoveryMetrics, setRecoveryMetrics] = useState(null);
  const [detectionMetrics, setDetectionMetrics] = useState(null);
  const [busy, setBusy] = useState('');
  const [error, setError] = useState('');
  const [logs, setLogs] = useState(['SYSTEM INITIALIZED. SELECT OR DROP AN IMAGE CANVAS.']);
  const [zoomTarget, setZoomTarget] = useState(null);

  const inputRef = useRef(null);
  const logRef = useRef(null);

  useEffect(() => {
    if (!file) {
      setPreview(null);
      return;
    }
    let active = true;
    const url = URL.createObjectURL(file);
    const image = new Image();
    image.onload = () => {
      if (!active) return;
      const canvas = document.createElement('canvas');
      canvas.width = image.naturalWidth;
      canvas.height = image.naturalHeight;
      const context = canvas.getContext('2d');
      if (context) {
        context.drawImage(image, 0, 0);
        setPreview(canvas.toDataURL('image/png'));
      } else {
        setPreview(url);
      }
    };
    image.onerror = () => {
      if (active) setError('Could not decode the selected image file.');
    };
    image.src = url;
    return () => {
      active = false;
      URL.revokeObjectURL(url);
    };
  }, [file]);

  useEffect(() => {
    if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight;
  }, [logs]);

  // Handle escape to close zoom
  useEffect(() => {
    const handleKeyDown = e => {
      if (e.key === 'Escape') setZoomTarget(null);
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const addLog = msg => {
    const time = new Date().toLocaleTimeString('en-US', { hour12: false });
    setLogs(prev => [...prev, `${time} // ${msg}`]);
  };

  const acceptFile = nextFile => {
    if (!nextFile) return;
    if (!['image/png', 'image/jpeg'].includes(nextFile.type)) {
      setError('Invalid format: Select a PNG or JPEG file.');
      return;
    }
    setFile(nextFile);
    setPreview(null);
    setImages(initialImages);
    setEmbedMetrics(null);
    setVerifyReport(null);
    setRecoveryMetrics(null);
    setDetectionMetrics(null);
    setError('');
    addLog(`LOADED ASSET: ${nextFile.name} (${Math.round(nextFile.size / 1024)} KB)`);
  };

  const run = async (actionName, callback) => {
    if (busy) return;
    setBusy(actionName);
    setError('');
    addLog(`INITIATING: ${actionName.toUpperCase()}`);
    try {
      await callback();
      addLog(`COMPLETED: ${actionName.toUpperCase()}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
      addLog(`ABORTED [ERROR]: ${msg}`);
    } finally {
      setBusy('');
    }
  };

  const regionFields = () => {
    const values = Object.fromEntries(
      ['x', 'y', 'w', 'h'].map(name => [name, Number(region[name])])
    );
    if (
      ['x', 'y', 'w', 'h'].some(name => region[name] === '' || !Number.isSafeInteger(values[name])) ||
      values.x < 0 ||
      values.y < 0 ||
      values.w < 1 ||
      values.h < 1
    ) {
      throw new Error('Tamper coordinates require integers with X/Y >= 0 and W/H >= 1.');
    }
    return { ...values, mode: region.mode };
  };

  const embed = () =>
    run('Watermark Injection', async () => {
      const result = await postStage('embed', { image: file, key });
      setImages(prev => ({ ...initialImages, original: prev.original, watermarked: result.watermarked }));
      setEmbedMetrics({ psnr: result.psnr, ssim: result.ssim });
      setVerifyReport(null);
      setRecoveryMetrics(null);
      setDetectionMetrics(null);
      addLog('PAYLOAD EMBEDDED VIA 2-BIT LSB FRAGILE MATRIX.');
    });

  const tamper = () =>
    run('Tamper Simulation', async () => {
      const result = await postStage('tamper', {
        image: dataUrlToFile(images.watermarked, 'watermarked.png'),
        ...regionFields(),
      });
      setImages(prev => ({ ...prev, tampered: result.tampered, overlay: null, recovered: null }));
      setVerifyReport(null);
      setRecoveryMetrics(null);
      setDetectionMetrics(null);
      addLog(`SECTOR ALTERATION APPLIED: [${region.mode.toUpperCase()}] AT X:${region.x} Y:${region.y} W:${region.w} H:${region.h}`);
    });

  const verify = () =>
    run('Integrity Verification', async () => {
      const result = await postStage('verify', {
        image: dataUrlToFile(images.tampered, 'tampered.png'),
        key,
      });
      setImages(prev => ({ ...prev, overlay: result.overlay, recovered: null }));
      setVerifyReport(result.report);
      setRecoveryMetrics(null);
      addLog(`VERIFIER SCANNED: ${result.report?.flaggedBlocks ?? 0} COMPROMISED BLOCKS OF ${result.report?.totalBlocks ?? 0}`);
    });

  const recover = () =>
    run('Content Reconstruction', async () => {
      const result = await postStage('recover', {
        image: dataUrlToFile(images.tampered, 'tampered.png'),
        original: file,
        key,
      });
      setImages(prev => ({ ...prev, recovered: result.recovered }));
      setRecoveryMetrics({
        ...result.metrics,
        recoveredBlocks: result.report?.recoveredBlocks,
        unrecoveredBlocks: result.report?.unrecoveredBlocks,
      });
      addLog(`RECONSTRUCTION FINISHED: ${result.report?.recoveredBlocks ?? 0} BLOCKS RESTORED.`);
    });

  const analyze = () =>
    run('Full Pipeline Analysis', async () => {
      const result = await postStage('analyze', { image: file, key, ...regionFields() });
      setImages({
        original: result.original || preview,
        watermarked: result.watermarked,
        tampered: result.tampered,
        overlay: result.overlay,
        recovered: result.recovered,
      });
      setEmbedMetrics(result.imperceptibility);
      setDetectionMetrics(result.detection);
      setVerifyReport(
        result.report?.reportBefore
          ? { ...result.report.reportBefore, tamperMap: result.report.tamperMapBefore }
          : null
      );
      setRecoveryMetrics(result.recovery);
      addLog('ALL 5 STAGES EXECUTED. DETECTION & RECOVERY TELEMETRY SYNCHRONIZED.');
    });

  // Preset quick coordinates helper
  const applyPreset = preset => {
    if (preset === 'center') {
      setRegion(prev => ({ ...prev, x: '48', y: '48', w: '64', h: '64' }));
      addLog('PRESET APPLIED: CENTER SECTOR (64x64)');
    } else if (preset === 'corner') {
      setRegion(prev => ({ ...prev, x: '0', y: '0', w: '48', h: '48' }));
      addLog('PRESET APPLIED: TOP-LEFT SECTOR (48x48)');
    } else if (preset === 'strip') {
      setRegion(prev => ({ ...prev, x: '0', y: '50', w: '128', h: '24' }));
      addLog('PRESET APPLIED: HORIZONTAL STRIP (128x24)');
    }
  };

  const clearLogs = () => {
    setLogs(['TERMINAL CLEARED. STANDBY FOR DISPATCH.']);
  };

  return (
    <div className="relative min-h-screen text-comic-paper">
      {/* Background WebGL Shader */}
      <ShaderBg />

      {/* Main Container */}
      <div className="relative mx-auto max-w-[1520px] px-3 py-4 sm:px-6 sm:py-6 lg:px-8">
        {/* COMIC MASTHEAD HEADER */}
        <header className="mb-6 rounded-2xl border-2 border-comic-border bg-comic-panel p-4 shadow-comic sm:p-6">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            {/* Left Title & Issue Badge */}
            <div className="space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                <span className="inline-flex items-center gap-1.5 rounded border border-comic-red bg-comic-red px-2.5 py-0.5 font-comic text-xs uppercase tracking-wider text-white shadow-comic-sm">
                  <SpiderIcon className="w-3.5 h-3.5" />
                  ISSUE #01
                </span>
                <span className="rounded border border-comic-border bg-comic-card px-2.5 py-0.5 font-mono text-[11px] uppercase tracking-widest text-comic-blue">
                  SECTOR: LSB-FRAGILE MATRIX
                </span>
                <span className="rounded border border-comic-border bg-comic-card px-2.5 py-0.5 font-mono text-[11px] uppercase tracking-widest text-comic-yellow">
                  AUTHENTICITY ENGINE
                </span>
              </div>

              <div className="flex items-baseline gap-2">
                <h1 className="font-comic text-4xl tracking-wide text-white sm:text-5xl lg:text-6xl">
                  STEGNO<span className="text-comic-red">.</span>LAB
                </h1>
                <span className="font-display text-xs font-bold uppercase tracking-widest text-comic-muted sm:text-sm">
                  // SPIDER-SENSE INTEGRITY SUITE
                </span>
              </div>

              <p className="max-w-3xl font-sans text-xs leading-relaxed text-comic-muted sm:text-sm">
                High-fidelity fragile steganography workbench. Embed cryptographic LSB watermarks,
                simulate sector tampering, identify altered pixel blocks via hash parity, and reconstruct
                compromised image data.
              </p>
            </div>

            {/* Right Meta Info Badges */}
            <div className="flex flex-wrap items-center gap-2 self-start lg:flex-col lg:items-end">
              <div className="flex items-center gap-2 rounded-lg border-2 border-comic-border bg-comic-ink px-3 py-1.5 font-mono text-xs">
                <span className="h-2 w-2 rounded-full bg-comic-blue animate-pulse" />
                <span className="text-comic-muted">ENDPOINT:</span>
                <span className="font-bold text-comic-blue">/api</span>
              </div>
              <div className="rounded border border-comic-border bg-comic-card px-3 py-1 font-mono text-[11px] text-comic-muted">
                ACTIVE KEY: <span className="font-bold text-comic-paper">{key ? 'SET' : 'NONE'}</span>
              </div>
            </div>
          </div>
        </header>

        {/* Global Error Banner */}
        {error && (
          <div
            role="alert"
            className="mb-6 flex items-start justify-between gap-3 rounded-xl border-2 border-comic-red bg-comic-red/15 p-4 text-comic-paper shadow-comic-red"
          >
            <div className="flex items-center gap-3">
              <AlertIcon className="w-5 h-5 shrink-0 text-comic-red" />
              <div>
                <strong className="font-display uppercase tracking-wider text-comic-red">
                  Pipeline Alert:
                </strong>{' '}
                <span className="font-sans text-sm">{error}</span>
              </div>
            </div>
            <button
              onClick={() => setError('')}
              className="rounded border border-comic-red/40 px-2 py-0.5 font-mono text-xs font-bold text-comic-red hover:bg-comic-red hover:text-white"
              aria-label="Dismiss error"
            >
              DISMISS
            </button>
          </div>
        )}

        {/* Global Busy Status Indicator */}
        {busy && (
          <div
            role="status"
            aria-live="polite"
            className="mb-6 flex items-center gap-3 rounded-xl border-2 border-comic-blue bg-comic-blue/10 p-4 text-comic-paper shadow-comic-blue"
          >
            <div className="h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-comic-blue border-t-transparent" />
            <div className="font-display text-sm font-bold uppercase tracking-wider text-comic-blue">
              EXECUTING: {busy}…
            </div>
            <span className="hidden font-mono text-xs text-comic-muted sm:inline">
              (Processing high-resolution bitmap buffers)
            </span>
          </div>
        )}

        {/* MAIN WORKBENCH GRID: Responsive layout */}
        <div className="grid gap-6 xl:grid-cols-[380px_minmax(0,1fr)]">
          {/* LEFT SIDEBAR: Controls & Actions */}
          <aside className="space-y-6">
            {/* PANEL 01: Image Setup */}
            <section className="rounded-xl border-2 border-comic-border bg-comic-panel p-4 shadow-comic sm:p-5">
              <div className="mb-4 flex items-center justify-between border-b-2 border-comic-border pb-3">
                <div className="flex items-center gap-2">
                  <span className="rounded bg-comic-red px-2 py-0.5 font-comic text-xs tracking-wider text-white">
                    01
                  </span>
                  <h2 className="font-display text-sm font-bold uppercase tracking-wider text-comic-paper">
                    Source Ingestion
                  </h2>
                </div>
                {file && (
                  <span className="font-mono text-[10px] text-comic-blue">
                    READY
                  </span>
                )}
              </div>

              {/* Upload Drop Zone */}
              <div
                onDragOver={e => e.preventDefault()}
                onDrop={e => {
                  e.preventDefault();
                  acceptFile(e.dataTransfer.files[0]);
                }}
                onClick={() => inputRef.current?.click()}
                onKeyDown={e => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    inputRef.current?.click();
                  }
                }}
                tabIndex={0}
                role="button"
                aria-label="Upload image"
                className="group relative cursor-pointer overflow-hidden rounded-xl border-2 border-dashed border-comic-border bg-comic-ink/90 p-5 text-center transition-all hover:border-comic-red hover:bg-comic-ink focus:outline-none focus:ring-2 focus:ring-comic-red bg-halftone-dots"
              >
                <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-lg border-2 border-comic-border bg-comic-panel group-hover:border-comic-red group-hover:text-comic-red">
                  <UploadIcon className="w-6 h-6 text-comic-blue group-hover:text-comic-red transition-colors" />
                </div>
                <p className="font-display text-xs font-bold uppercase tracking-wider text-comic-paper">
                  {file ? file.name : 'Select or Drop Canvas'}
                </p>
                <p className="mt-1 font-mono text-[11px] text-comic-muted">
                  PNG OR JPEG FORMATS SUPPORTED
                </p>
                <input
                  ref={inputRef}
                  className="sr-only"
                  type="file"
                  accept="image/png,image/jpeg"
                  onChange={e => {
                    acceptFile(e.target.files[0]);
                    e.target.value = '';
                  }}
                  tabIndex={-1}
                />
              </div>

              {/* Secret Key Input */}
              <div className="mt-5">
                <div className="flex items-center justify-between">
                  <label
                    htmlFor="secret-key"
                    className="flex items-center gap-1.5 font-display text-xs font-bold uppercase tracking-wider text-comic-paper"
                  >
                    <LockIcon className="text-comic-blue" />
                    Cryptographic Passkey
                  </label>
                  <button
                    type="button"
                    onClick={() => setShowKey(!showKey)}
                    className="flex items-center gap-1 font-mono text-[10px] text-comic-muted hover:text-comic-blue"
                  >
                    {showKey ? <EyeOffIcon /> : <EyeIcon />}
                    <span>{showKey ? 'HIDE' : 'SHOW'}</span>
                  </button>
                </div>
                <input
                  id="secret-key"
                  className="mt-2 w-full rounded-lg border-2 border-comic-border bg-comic-ink px-3 py-2 font-mono text-sm text-comic-paper outline-none transition focus:border-comic-blue focus:ring-1 focus:ring-comic-blue"
                  type={showKey ? 'text' : 'password'}
                  value={key}
                  onChange={e => setKey(e.target.value)}
                  autoComplete="off"
                  spellCheck={false}
                  placeholder="Enter secret key…"
                />
                <p className="mt-1 font-mono text-[10px] text-comic-muted">
                  Key seeds the pseudo-random fragile hash mapping.
                </p>
              </div>

              {/* Tamper Coordinates & Mode */}
              <div className="mt-5 border-t-2 border-comic-border pt-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5 font-display text-xs font-bold uppercase tracking-wider text-comic-paper">
                    <TargetIcon className="text-comic-red" />
                    Tamper Coordinates (PX)
                  </div>
                  {/* Quick Presets */}
                  <div className="flex items-center gap-1">
                    <button
                      type="button"
                      onClick={() => applyPreset('center')}
                      className="rounded border border-comic-border bg-comic-ink px-1.5 py-0.5 font-mono text-[9px] uppercase text-comic-muted hover:border-comic-red hover:text-comic-paper"
                    >
                      Center
                    </button>
                    <button
                      type="button"
                      onClick={() => applyPreset('corner')}
                      className="rounded border border-comic-border bg-comic-ink px-1.5 py-0.5 font-mono text-[9px] uppercase text-comic-muted hover:border-comic-red hover:text-comic-paper"
                    >
                      Corner
                    </button>
                    <button
                      type="button"
                      onClick={() => applyPreset('strip')}
                      className="rounded border border-comic-border bg-comic-ink px-1.5 py-0.5 font-mono text-[9px] uppercase text-comic-muted hover:border-comic-red hover:text-comic-paper"
                    >
                      Strip
                    </button>
                  </div>
                </div>

                <div className="mt-2.5 grid grid-cols-4 gap-2">
                  {['x', 'y', 'w', 'h'].map(coord => (
                    <label key={coord} className="font-mono text-[11px] text-comic-muted">
                      <span className="uppercase font-bold text-comic-paper">{coord}</span>
                      <input
                        className="mt-1 w-full rounded border-2 border-comic-border bg-comic-ink px-2 py-1.5 font-mono text-xs text-comic-paper outline-none focus:border-comic-red"
                        type="number"
                        min={coord === 'w' || coord === 'h' ? 1 : 0}
                        step="1"
                        value={region[coord]}
                        onChange={e =>
                          setRegion(prev => ({ ...prev, [coord]: e.target.value }))
                        }
                      />
                    </label>
                  ))}
                </div>

                <div className="mt-3">
                  <label
                    htmlFor="tamper-mode"
                    className="block font-mono text-[11px] uppercase tracking-wider text-comic-muted"
                  >
                    Tamper Mutation Mode
                  </label>
                  <select
                    id="tamper-mode"
                    className="mt-1 w-full rounded-lg border-2 border-comic-border bg-comic-ink px-3 py-2 font-display text-xs font-semibold text-comic-paper outline-none focus:border-comic-red"
                    value={region.mode}
                    onChange={e => setRegion(prev => ({ ...prev, mode: e.target.value }))}
                  >
                    <option value="black">Black Fill (Constant Zero)</option>
                    <option value="white">White Fill (Constant 255)</option>
                    <option value="noise">Random Noise (Entropy Injection)</option>
                    <option value="copy">Copy Region (Displacement Clone)</option>
                  </select>
                </div>
              </div>
            </section>

            {/* PANEL 02: Execution Actions */}
            <section className="rounded-xl border-2 border-comic-border bg-comic-panel p-4 shadow-comic sm:p-5">
              <div className="mb-4 flex items-center justify-between border-b-2 border-comic-border pb-3">
                <div className="flex items-center gap-2">
                  <span className="rounded bg-comic-blue px-2 py-0.5 font-comic text-xs tracking-wider text-comic-ink">
                    02
                  </span>
                  <h2 className="font-display text-sm font-bold uppercase tracking-wider text-comic-paper">
                    Protocol Workflow
                  </h2>
                </div>
                <span className="font-mono text-[10px] text-comic-muted">
                  STEP-BY-STEP
                </span>
              </div>

              {/* Step Buttons */}
              <div className="space-y-2">
                <button
                  type="button"
                  disabled={!file || !!busy}
                  onClick={embed}
                  className="flex w-full items-center justify-between rounded-lg border-2 border-comic-border bg-comic-card px-3.5 py-2.5 font-display text-xs font-bold uppercase tracking-wider text-comic-paper shadow-comic-sm transition-all hover:border-comic-blue hover:text-comic-blue active:translate-x-0.5 active:translate-y-0.5 active:shadow-none disabled:cursor-not-allowed disabled:border-comic-border/50 disabled:bg-comic-ink/50 disabled:text-comic-muted/40"
                >
                  <span className="flex items-center gap-2">
                    <span className="font-mono text-comic-blue">02 //</span>
                    WATERMARK INJECTION
                  </span>
                  <span className="font-mono text-[10px] text-comic-muted">[EMBED]</span>
                </button>

                <button
                  type="button"
                  disabled={!images.watermarked || !!busy}
                  onClick={tamper}
                  className="flex w-full items-center justify-between rounded-lg border-2 border-comic-border bg-comic-card px-3.5 py-2.5 font-display text-xs font-bold uppercase tracking-wider text-comic-paper shadow-comic-sm transition-all hover:border-comic-red hover:text-comic-red active:translate-x-0.5 active:translate-y-0.5 active:shadow-none disabled:cursor-not-allowed disabled:border-comic-border/50 disabled:bg-comic-ink/50 disabled:text-comic-muted/40"
                >
                  <span className="flex items-center gap-2">
                    <span className="font-mono text-comic-red">03 //</span>
                    APPLY TAMPER MUTATION
                  </span>
                  <span className="font-mono text-[10px] text-comic-muted">[TAMPER]</span>
                </button>

                <button
                  type="button"
                  disabled={!images.tampered || !!busy}
                  onClick={verify}
                  className="flex w-full items-center justify-between rounded-lg border-2 border-comic-border bg-comic-card px-3.5 py-2.5 font-display text-xs font-bold uppercase tracking-wider text-comic-paper shadow-comic-sm transition-all hover:border-comic-yellow hover:text-comic-yellow active:translate-x-0.5 active:translate-y-0.5 active:shadow-none disabled:cursor-not-allowed disabled:border-comic-border/50 disabled:bg-comic-ink/50 disabled:text-comic-muted/40"
                >
                  <span className="flex items-center gap-2">
                    <span className="font-mono text-comic-yellow">04 //</span>
                    VERIFY TAMPER INTEGRITY
                  </span>
                  <span className="font-mono text-[10px] text-comic-muted">[VERIFY]</span>
                </button>

                <button
                  type="button"
                  disabled={!images.tampered || !!busy}
                  onClick={recover}
                  className="flex w-full items-center justify-between rounded-lg border-2 border-comic-border bg-comic-card px-3.5 py-2.5 font-display text-xs font-bold uppercase tracking-wider text-comic-paper shadow-comic-sm transition-all hover:border-comic-blue hover:text-comic-blue active:translate-x-0.5 active:translate-y-0.5 active:shadow-none disabled:cursor-not-allowed disabled:border-comic-border/50 disabled:bg-comic-ink/50 disabled:text-comic-muted/40"
                >
                  <span className="flex items-center gap-2">
                    <span className="font-mono text-comic-blue">05 //</span>
                    RECONSTRUCT CONTENT
                  </span>
                  <span className="font-mono text-[10px] text-comic-muted">[RECOVER]</span>
                </button>
              </div>

              {/* Comic Separator */}
              <div className="my-4 flex items-center gap-3">
                <span className="h-0.5 flex-1 bg-comic-border" />
                <span className="font-comic text-xs uppercase tracking-widest text-comic-muted">
                  OR MASTER SEQUENCE
                </span>
                <span className="h-0.5 flex-1 bg-comic-border" />
              </div>

              {/* Master Full Pipeline Button */}
              <button
                type="button"
                disabled={!file || !!busy}
                onClick={analyze}
                className="group relative flex w-full items-center justify-center gap-2 overflow-hidden rounded-xl border-2 border-comic-red bg-comic-red px-4 py-3.5 font-display text-sm font-black uppercase tracking-wider text-white shadow-comic-red transition-all hover:bg-comic-redDark active:translate-x-0.5 active:translate-y-0.5 active:shadow-none disabled:cursor-not-allowed disabled:border-comic-border disabled:bg-comic-border disabled:text-comic-muted/50 disabled:shadow-none"
              >
                <CrosshairBoltIcon className="w-5 h-5 transition-transform group-hover:scale-110" />
                <span>EXECUTE FULL PIPELINE</span>
              </button>
              <p className="mt-2 text-center font-mono text-[10px] text-comic-muted">
                Executes all 5 stages in a single unified multipart transaction.
              </p>
            </section>

            {/* PANEL 03: Dispatch Log Console */}
            <section className="overflow-hidden rounded-xl border-2 border-comic-border bg-comic-panel shadow-comic">
              <div className="flex items-center justify-between border-b-2 border-comic-border bg-comic-panelAlt px-4 py-2.5">
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-comic-blue animate-pulse" />
                  <h3 className="font-display text-xs font-bold uppercase tracking-wider text-comic-paper">
                    Dispatch Teletype
                  </h3>
                </div>
                <button
                  type="button"
                  onClick={clearLogs}
                  className="font-mono text-[10px] text-comic-muted hover:text-comic-red"
                >
                  CLEAR
                </button>
              </div>
              <div
                ref={logRef}
                role="log"
                aria-live="polite"
                className="h-44 space-y-1.5 overflow-y-auto bg-comic-ink p-3 font-mono text-[11px] leading-relaxed text-comic-muted"
              >
                {logs.map((entry, idx) => (
                  <div key={idx} className="flex items-start gap-2">
                    <span className="select-none text-comic-red">»</span>
                    <span className="break-all">{entry}</span>
                  </div>
                ))}
              </div>
            </section>
          </aside>

          {/* RIGHT MAIN VIEWPORT: Telemetry, Stage Images & Tamper Radar */}
          <main className="min-w-0 space-y-6">
            {/* TELEMETRY DECK */}
            <section className="rounded-xl border-2 border-comic-border bg-comic-panel p-4 shadow-comic sm:p-5">
              <div className="mb-4 flex flex-wrap items-center justify-between gap-2 border-b-2 border-comic-border pb-3">
                <div>
                  <h2 className="font-display text-sm font-bold uppercase tracking-wider text-comic-paper">
                    Telemetry & Metric Diagnostics
                  </h2>
                  <p className="font-mono text-xs text-comic-muted">
                    Signal-to-noise ratio, perceptual structural similarity, and block detection rates.
                  </p>
                </div>
                <span className="rounded bg-comic-card border border-comic-border px-2.5 py-1 font-mono text-[10px] font-bold text-comic-blue">
                  PARITY // LIVE
                </span>
              </div>

              <div className="space-y-4">
                {/* Section A: Imperceptibility */}
                <div>
                  <div className="mb-2 flex items-center gap-2">
                    <span className="h-1.5 w-1.5 rounded-full bg-comic-blue" />
                    <span className="font-display text-xs font-bold uppercase tracking-wider text-comic-muted">
                      Watermark Fidelity (Imperceptibility)
                    </span>
                  </div>
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
                    <ComicMetric
                      label="PSNR"
                      value={embedMetrics?.psnr}
                      unit="dB"
                      status={embedMetrics?.psnr ? (embedMetrics.psnr > 40 ? 'OPTIMAL' : 'ACCEPTABLE') : null}
                    />
                    <ComicMetric
                      label="SSIM"
                      value={embedMetrics?.ssim}
                      status={embedMetrics?.ssim ? (embedMetrics.ssim > 0.98 ? 'HIGH FIDELITY' : 'MODERATE') : null}
                    />
                    <ComicMetric
                      label="Target Mode"
                      value="LSB-2"
                      status="DUAL BIT"
                    />
                  </div>
                </div>

                {/* Section B: Tamper Detection */}
                <div>
                  <div className="mb-2 flex items-center gap-2">
                    <span className="h-1.5 w-1.5 rounded-full bg-comic-red" />
                    <span className="font-display text-xs font-bold uppercase tracking-wider text-comic-muted">
                      Tamper Radar (Detection Accuracy)
                    </span>
                  </div>
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                    <ComicMetric
                      label="Flagged / Total"
                      value={
                        (detectionMetrics?.flagged ?? verifyReport?.flaggedBlocks) != null
                          ? `${detectionMetrics?.flagged ?? verifyReport.flaggedBlocks} / ${
                              detectionMetrics?.total ?? verifyReport.totalBlocks
                            }`
                          : null
                      }
                      status={
                        (detectionMetrics?.flagged ?? verifyReport?.flaggedBlocks) ? 'ANOMALY FOUND' : null
                      }
                    />
                    <ComicMetric
                      label="True Positive (TPR)"
                      value={detectionMetrics?.tpr != null ? detectionMetrics.tpr * 100 : null}
                      unit="%"
                    />
                    <ComicMetric
                      label="False Positive (FPR)"
                      value={detectionMetrics?.fpr != null ? detectionMetrics.fpr * 100 : null}
                      unit="%"
                    />
                    <ComicMetric
                      label="Ground Truth Blocks"
                      value={detectionMetrics?.groundTruthBlocks}
                    />
                  </div>
                </div>

                {/* Section C: Content Recovery */}
                <div>
                  <div className="mb-2 flex items-center gap-2">
                    <span className="h-1.5 w-1.5 rounded-full bg-comic-yellow" />
                    <span className="font-display text-xs font-bold uppercase tracking-wider text-comic-muted">
                      Content Reconstruction (Recovery Quality)
                    </span>
                  </div>
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                    <ComicMetric
                      label="Recovery PSNR"
                      value={recoveryMetrics?.psnr}
                      unit="dB"
                    />
                    <ComicMetric
                      label="Recovery SSIM"
                      value={recoveryMetrics?.ssim}
                    />
                    <ComicMetric
                      label="Recovered Blocks"
                      value={recoveryMetrics?.recoveredBlocks}
                      status={recoveryMetrics?.recoveredBlocks ? 'RESTORED' : null}
                    />
                    <ComicMetric
                      label="Unrecovered Blocks"
                      value={recoveryMetrics?.unrecoveredBlocks}
                    />
                  </div>
                </div>
              </div>
            </section>

            {/* 5-STAGE COMIC PANEL STRIP */}
            <div>
              <div className="mb-3 flex items-center justify-between">
                <h3 className="font-comic text-xl tracking-wider text-comic-paper">
                  PIPELINE CANVASES // 5-STAGE STRIP
                </h3>
                <span className="font-mono text-xs text-comic-muted">
                  CLICK ANY PANEL TO ENLARGE
                </span>
              </div>

              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                <ComicStageImage
                  step="01"
                  label="Original Asset"
                  src={images.original || preview}
                  hint="Select or drag an image canvas into panel 01."
                  onZoom={setZoomTarget}
                />
                <ComicStageImage
                  step="02"
                  label="Watermarked Matrix"
                  src={images.watermarked}
                  hint="Run Watermark Injection to generate the stego carrier."
                  onZoom={setZoomTarget}
                />
                <ComicStageImage
                  step="03"
                  label="Tampered Asset"
                  src={images.tampered}
                  hint="Apply Tamper Mutation to simulate malicious edits."
                  onZoom={setZoomTarget}
                />
                <ComicStageImage
                  step="04"
                  label="Verification Overlay"
                  src={images.overlay}
                  hint="Verify Tamper Integrity to visualize flagged blocks."
                  onZoom={setZoomTarget}
                />
                <ComicStageImage
                  step="05"
                  label="Reconstructed Canvas"
                  src={images.recovered}
                  hint="Run Content Reconstruction to restore altered blocks."
                  onZoom={setZoomTarget}
                />
              </div>
            </div>

            {/* TAMPER RADAR GRID CANVAS */}
            <TamperGrid map={verifyReport?.tamperMap} report={verifyReport} />
          </main>
        </div>

        {/* COMIC FOOTER STRIP */}
        <footer className="mt-12 rounded-xl border-2 border-comic-border bg-comic-panel p-4 shadow-comic">
          <div className="flex flex-col items-center justify-between gap-3 text-center sm:flex-row sm:text-left">
            <div className="flex items-center gap-2">
              <SpiderIcon className="w-4 h-4 text-comic-red" />
              <span className="font-comic text-sm tracking-wider text-comic-paper">
                STEGNO CRYPTOGRAPHIC LABORATORY
              </span>
              <span className="font-mono text-xs text-comic-muted">
                // REVAMP SPEC 2026
              </span>
            </div>
            <div className="font-mono text-xs text-comic-muted">
              LSB-2 MATRIX ENGINE · NO EXTERNAL TRACKERS · HARDWARE ACCELERATED SHADER
            </div>
          </div>
        </footer>
      </div>

      {/* Image Zoom Modal */}
      {zoomTarget && (
        <ImageModal
          image={zoomTarget}
          label="Stage Inspection"
          onClose={() => setZoomTarget(null)}
        />
      )}
    </div>
  );
}

createRoot(document.getElementById('root')).render(<App />);
