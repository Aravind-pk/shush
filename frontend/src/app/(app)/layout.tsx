import { AppSidebar } from "@/components/app-sidebar";
import { AppChromeProvider, AppTopbar } from "@/components/app-chrome";
import { TooltipProvider } from "@/components/ui/tooltip";

export default function AppLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <TooltipProvider>
      <AppChromeProvider>
        <div className="flex min-h-screen w-full bg-[#0a0a0b] text-[#e9e9ec]">
          <AppSidebar />

          <main className="flex min-w-0 flex-1 flex-col">
            <AppTopbar />
            <div className="flex w-full max-w-[1460px] flex-col gap-[22px] px-[30px] pt-[26px] pb-[60px]">
              {children}
            </div>
          </main>
        </div>
      </AppChromeProvider>
    </TooltipProvider>
  );
}
