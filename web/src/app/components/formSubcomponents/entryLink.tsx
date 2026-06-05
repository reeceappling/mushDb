import {
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {JSX, ReactNode} from "react";
import {BaseExternalUrl} from "@/app/components/Constants";
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
export function EntryLinkForEntry<T extends Entry>(
    {
        props,
    }: {
        props: {
            entry: T;
            openInNewTab?: boolean;
        };
    }) {
    return <EntryLinkWrapper props={props}>
        {props.entry.getId()/* TODO: wrap in text tag?*/}
    </EntryLinkWrapper>
}

// export function DefaultLink<T extends Entry>(info:{
//     entry:T,
//     openInNewTab?: boolean
// }): JSX.Element { // TODO: consider using???
//     return <EntryLinkForEntry props={info}/>
// }
// export function DefaultIdLink(info:{
//     entryType: string,
//     linkId: string,
//     openInNewTab?: boolean,
//     displayId?: string
// }): JSX.Element { // TODO: consider using???
//     return <EntryLinkForId props={info}/>
// }

// export function EntryLinkWrapper<T extends Entry | {
//     linkId: string;
//     entryType: string;
// }>(
//     {props,children}:{
//         props:{data:T,openInNewTab?:boolean};
//         children: ReactNode;
//     }){
//     if ("linkId" in props.data) {
//         return <EntryLinkWrapperForId props={{
//             linkId: props.data.linkId,
//             entryType: props.data.entryType,
//             openInNewTab: props.openInNewTab,
//         }}>
//             {children}
//         </EntryLinkWrapperForId>
//     } else {
//         return <EntryLinkWrapperForEntry props={{
//             entry: props.data,
//             openInNewTab:true,
//         }}>
//             {children}
//         </EntryLinkWrapperForEntry>
//     }
// }

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