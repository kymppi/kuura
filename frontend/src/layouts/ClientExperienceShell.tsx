import { Notification } from '@carbon/icons-react';
import {
  Content,
  Header,
  HeaderContainer,
  HeaderGlobalAction,
  HeaderGlobalBar,
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
                <HeaderGlobalBar>
                  <HeaderGlobalAction
                    aria-label="Notifications"
                    onClick={() => alert('notification click')}
                  >
                    <Notification size={20} />
                  </HeaderGlobalAction>
                </HeaderGlobalBar>
              </Header>
            </>
          )}
        />
      </Theme>

      <Content>{children}</Content>
    </>
  );
}
