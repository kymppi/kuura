import {
  Content,
  Header,
  HeaderContainer,
  HeaderMenuButton,
  HeaderMenuItem,
  HeaderName,
  HeaderNavigation,
  HeaderSideNavItems,
  SideNav,
  SideNavItems,
  SkipToContent,
  Theme,
} from '@carbon/react';
import type React from 'react';

const MenuItems = ({ activePath }: { activePath: string }) => (
  <>
    <HeaderMenuItem href="/account" isActive={activePath === '/account'}>
      My account
    </HeaderMenuItem>
    <HeaderMenuItem href="/sessions" isActive={activePath === '/sessions'}>
      Current sessions
    </HeaderMenuItem>
  </>
);

export default function ClientExperienceShell({
  children,
  activePath,
  layoutName,
}: {
  children: React.ReactNode;
  activePath: string;
  layoutName: string;
}) {
  return (
    <>
      <Theme theme="g100">
        <HeaderContainer
          render={({ isSideNavExpanded, onClickSideNavExpand }) => (
            <>
              <Header aria-label="Header">
                <SkipToContent />
                <HeaderMenuButton
                  aria-label={isSideNavExpanded ? 'Close menu' : 'Open menu'}
                  onClick={onClickSideNavExpand}
                  isActive={isSideNavExpanded}
                  aria-expanded={isSideNavExpanded}
                />
                <HeaderName href="/" prefix="KUURA">
                  {layoutName}
                </HeaderName>
                <HeaderNavigation aria-label="Header Navigation">
                  <MenuItems activePath={activePath} />
                </HeaderNavigation>
                <SideNav
                  aria-label="Side navigation"
                  expanded={isSideNavExpanded}
                  isPersistent={false}
                  onSideNavBlur={onClickSideNavExpand}
                >
                  <SideNavItems>
                    <HeaderSideNavItems>
                      <MenuItems activePath={activePath} />
                    </HeaderSideNavItems>
                  </SideNavItems>
                </SideNav>
              </Header>
            </>
          )}
        />
      </Theme>

      <Content style={{ height: 'calc(100dvh - 48px)', padding: 0 }}>
        {children}
      </Content>
    </>
  );
}
