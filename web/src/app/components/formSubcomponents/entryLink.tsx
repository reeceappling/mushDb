import {
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {ReactNode} from "react";
import {BaseExternalUrl} from "@/app/components/Constants";
import {viewUrlFor} from "@/app/components/common";

export default function EntryLink(
    {
        props,
        children,
    }: {
        props: {
            displayedId: string;
            linkId: string;
            entryType: string;
            openInNewTab?: boolean;
        };
        children: ReactNode;
    }) {
    const {dispatch} = useRfidReaderContext()
    const doNewTab = () => {
        // TODO: THIS!!!!!
    }
    const changeModal = () => {
        dispatch({
            type: ActionTypes.SET_MODAL_INFO,
            payload: {modalType: props.entryType, recordId: props.displayedId} // TODO: IS THIS OK?
        })
    }
    return <div  data-cy-id="EntryLink" onClick={props.openInNewTab?doNewTab:changeModal}>{/* TODO: REMOVE MODAL??? */}
        {children}
    </div>
}

export function EntryLinkWrapper(
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