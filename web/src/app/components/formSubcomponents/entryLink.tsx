

import { ReactNode} from "react";
import {Entry, EntryUrlId, viewUrlFor} from "@/app/components/common";

export default function EntryLinkForId(
    {
        props,
    }: {
        props: {
            entryType: string,
            linkId: string,
            displayId?: string,
            openInNewTab?: boolean;
        };
    }) {
    return <EntryLinkIdWrapper props={props}>
        {props.displayId || props.linkId/* TODO: wrap in text tag?*/}
    </EntryLinkIdWrapper>
}

export function EntryLinkWrapper<T extends Entry>(
    {
        props,
        children,
    }: {
        props: {
            entry: T;
            openInNewTab?: boolean;
        };
        children: ReactNode;
    }) {
    return <EntryLinkInternal props={{
        entryType: props.entry.entryType(),
        linkId: EntryUrlId(props.entry),
        openInNewTab: props.openInNewTab,
    }}>{children}</EntryLinkInternal>
}

// TODO: swap all of these over to EntryLinkWrapper???
export function EntryLinkIdWrapper(
    {
        props,
        children,
    }: {
        props: {
            linkId: string;
            entryType: string;
            openInNewTab?: boolean;
        };
        children: ReactNode;
    }) {
    const onClickStopPropagation = (e: React.MouseEvent) => {
        e.stopPropagation();
    }
    const actualLink = viewUrlFor(props.entryType, props.linkId)
    if (props.openInNewTab===true){
        return <a href={actualLink} target={"_blank"} rel={"noopener noreferrer"} onClick={onClickStopPropagation}>
            {children}
        </a>
    }
    return <a href={actualLink} onClick={onClickStopPropagation}>
        {children}
    </a>
}

export function EntryLinkInternal(
    {
        props,
        children,
    }: {
        props: {
            entryType: string,
            linkId: string,
            openInNewTab?: boolean;
        };
        children: ReactNode;
    }) {
    const actualLink = viewUrlFor(props.entryType, props.linkId)
    if (props.openInNewTab===true){
        return <a href={actualLink} target={"_blank"} rel={"noopener noreferrer"} onClick={e=>e.stopPropagation()}>
            {children}
        </a>
    }
    return <a href={actualLink} onClick={e=>e.stopPropagation()}>
        {children}
    </a>
}