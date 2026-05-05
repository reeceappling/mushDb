'use client'

import React, {useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {BagData} from "@/app/components/bagServer";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {InnocDisplay, NewTransferArea, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    DisplayInput,
    DisposedSaleContamArea,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea, ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData,
    SendMultipartRequest,
    setFormData,
    setFormImages
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    GensInlineDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    NameArea,
    ParentDisplay,
    PicsDisplay,
    SpeciesArea,
    SubspeciesArea
} from "@/app/components/formSubcomponents/commonClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {CreatedLinkFor, SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState, InitialNotesState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import GenerationArea from "@/app/components/formSubcomponents/generationInput";
import {TopLevelImageSelector} from "@/app/components/formSubcomponents/imageSelector";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {PcRunData, RecentPCRunSelector} from "@/app/components/pcRunServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {NewFruitForm} from "@/app/components/fruitClient";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector, SubspeciesFormArea} from "@/app/components/subspeciesClient";
import {FruitData} from "@/app/components/fruitServer";
import {SubstrateBatchArea, SubstrateBatchSelector} from "@/app/components/substrateBatchClient";
import WetnessSlider from "@/app/components/formSubcomponents/utils/slider";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsQuadColArea, QuadColLastCol} from "@/app/components/pcRunClient";
import {TransferData} from "@/app/components/transferServer";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {InlineEntry} from "@/app/components/agarRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";

export function AssertBag(input: any): asserts input is BagData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['filterSize', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Bag assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['sealDate', 'number'],
        ['knownFruitable', 'boolean'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['sale', 'string'],
        ['substrateBatch', 'string'],
        ['disposed', 'number'],
        ['wetness', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexRequiredKeys = new Map<string, (v: any) => boolean>([
        // ['entryType', (inp: any) => {
        //     return (typeof inp === 'string' && inp === "bag")
        // }],
    ])
    for (let [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Bag assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['flushes', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Bag assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

// TODO: MOVE!
export function OvcForXfers(parentId: string, parentType: string, validTypesTo: string[], cookies: string, addTransferOut?: (xfer: TransferData) => void, altTxt?: string): OnViewCreatorQuadCol {
    return {
        txt: altTxt || "New Transfer",
        newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
            return <NewTransferArea cookies={cookies} idFrom={parentId} typeFrom={parentType} validTypesTo={validTypesTo}
                                    onCreated={(xfer: TransferData) => {
                                        addTransferOut && addTransferOut(xfer)
                                        onCreate([{
                                            typeText: "Transfer",
                                            node: <CreatedLinkFor linkId={xfer._id} typ={"transfer"}/>,
                                            lastNode: <QuadColLastCol dstType={xfer.toType} id={xfer.to}/>
                                        }])
                                    }}/>
        }
    }
}

// TODO: MOVE!
export function OvcForNewFruit(parentId: string, parentType: string, cookies: string): OnViewCreatorQuadCol {
    return {
        txt: "Record New Fruit",
        newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
            return <NewFruitForm parentId={parentId} parentType={parentType} readonly={false} cookies={cookies}
                                 onCreate={(fr: FruitData) => {
                                     onCreate([{
                                         typeText: "Fruit",
                                         node: <CreatedLinkFor linkId={fr._id} typ={"fruit"}/>,
                                     }])
                                 }}/>
        }
    }
}

export default function BagDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertBag(data)
        const [initial, setInitial] = useState(data)

        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [sale, setSale] = useState(initial.sale)
        const [transfersOut, setTransfersOut] = useState(initial.transfersOut || [])
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
        // ItemsWithPics
        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [flushes, setFlushes] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.flushes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        //const [newFruits, setNewFruits] = useState<FruitData[]>([]) // TODO: get rid of???
        const filterSizeArea = (filterSize: string, headerLevel?: number) => {
            return <div>
                <div>{"Filter Size: " + filterSize}</div>
            </div>
        }
        const updateInitial = (updated: BagData)=> {
            setInitial(updated)
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setTransfersOut(updated.transfersOut || [])
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setPics(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setFlushes(InitialPicsEntries(updated.flushes))
            setAcl(updated.acl)
        }
        const bagSubmit = () => {
            let body = new FormData()
            let dataObj: any = {
                knownFruitable: knownFruitable,
                sale: sale, // TODO: how/when should sales be made?
                notes: notes,
                disposed: disposed,
                writeTagTo: writeTagTo,
                acl: MarshalAcl(acl),
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Flushes
                let flushesInfo = resolvePicsFormData(flushes)
                let newFlushes = flushesInfo.images
                dataObj.flushes = flushesInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                //body.set("data", JSON.stringify(dataObj))
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
                setFormImages(body, "newFlush", newFlushes)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl + "/db/update/bag/" + initial._id, cookies, body)
                // fetch(BaseExternalUrl + "/db/update/bag/" + data._id, {
                //     method: 'Post',
                //     body: body,
                //     headers: {
                //         'Content-type': "multipart/form-data",
                //         credentials: 'include',
                //         //'Cookie': cookies,
                //     },
                // })
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    AssertBag(newEntry)
                    updateInitial(newEntry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            OvcForNewFruit(initial._id, "bag", cookies), // TODO: test heavily
        ]
        return (
            <DisplayFormWrapper entryType={"bag"}>
                {/* TODO: ok?<div className={"sectionHolder"}>*/}
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Bag"} entryType={"bag"}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
                <FlexedArea>{/* TODO: validate that this is working as intended*/}
                    <FlexedSinglesGroup>
                        <SubstrateRecipeArea id={data.recipe} readonly={true} txt={"Substrate recipe: "}/>
                        <SubstrateBatchArea id={data.substrateBatch} txt={"Substrate batch: "} readonly={true}/>
                        <SubstrateRecipeArea id={initial.recipe} headerLevel={headerLevel} readonly={true}/>
                        {filterSizeArea(initial.filterSize)}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <DateArea pre={"PC Date: "} readonly={true} when={initial.creationDate}/>
                        <DateArea pre={"Seal Date: "} readonly={true} when={initial.sealDate}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                        <DisposedDisplay disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <KnownFruitableArea initial={initial.knownFruitable} doSelect={setKnownFruitable}
                                            readonly={readonly}
                                            headerLevel={headerLevel}/>
                        <SaleArea sale={initial.sale} setSale={setSale} readonly={readonly}
                                  headerLevel={headerLevel} canCreateSale={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore}
                                         gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                        <WetnessDisplay value={initial.wetness}/>{/* TODO: FIX */}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType}
                                       headerLevel={headerLevel}/>
                        <InnocDisplay innoc={initial.innoc}/>
                    </FlexedSinglesGroup>
                </FlexedArea>

                <TransfersOutDisplay validTypesTo={["plate"]} thisId={initial._id} thisEntryType={"bag"}
                                     transfersOut={initial.transfersOut}
                                     allowNewTransferCreation={true}
                                     cookies={cookies}/>
                <PicsDisplay pix={pics} readonly={readonly} updateParent={setPics}/>{/* Pics */}
                {/* Flushes */}
                <PicsDisplay pix={flushes} readonly={readonly}
                             updateParent={setFlushes} addButtonText={"Create New Flush"}
                             sectionHeader={"Flushes: "}/>

                <ContamsDisplay initial={initial.contamination || []} current={contams} updateParent={setContams}
                                readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {/* Write tag area */}
                {readonly ? null : <ReaderWriterSelector txt={"Writer to write to: "} onSelect={setWriteTagTo}
                                                         headerLevel={headerLevel}/>}
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    bagSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Bag data format incorrect: " + err}</div>
    }
}

export function WetnessDisplay({value}: { value?: number }) {
    return <TestAndValidate todos={["fix"]}>
        <div>{"Wetness: " + (value ? value+"/10" : "unknown")}</div>
    </TestAndValidate>
}

const filterSizeSelector = (setFilterSize: (f?: string) => void, filterSize?: string) => {
    const possibleFilterSizes = ["5nm", "8nm", "unknown"]
    return <div className={"centerH medGapTop"}>
        {"Filter size: "}<select className={"tailwindSelector"} value={filterSize || ""} onChange={
        (evt) => {
            setFilterSize(evt.currentTarget.value === "" ? undefined : evt.currentTarget.value)
        }
    }>
        {possibleFilterSizes.map(function (size, i) {
            return <option value={size} key={i}>{size}</option>
        })}
    </select></div>
}

export function NewBagForm({handlers, substrateBatchIn, pcRunIn}: {
    handlers: NewEntryInput<BagData>,
    pcRunIn?: PcRunData,
    substrateBatchIn?: SubstrateBatchData,
}) {
    // Required defined
    const [substrateBatch, setSubstrateBatch] = useState(substrateBatchIn)
    const [wetness, setWetness] = useState<number | undefined>(undefined)
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [filterSize, setFilterSize] = useState<string | undefined>()
    // Optional
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    // TODO: handle handlers.isTOpLevel
    const newBagSubmit = () => {
        if (pcRun === undefined) {
            setErr("PC Run cannot be undefined!");
            return
        }
        if (filterSize === undefined) {
            setErr("Filter Size cannot be undefined!");
            return
        }
        if (substrateBatch === undefined) {
            setErr("Substrate batch cannot be undefined!");
            return
        }
        let body: any = {
            pcRun: pcRun._id,
            filterSize: filterSize,
            //creationDate: creationDate, // TODO: REMOVED CREATION DATE!
            wetness: wetness,
            substrateBatch: substrateBatch._id,
            writeTagTo: writeTagTo,
            notes: notes,
        }
        fetch(BaseExternalUrl + "/create/bag", {
            method: 'Post',
            body: JSON.stringify(body),
            headers: {
                credentials: 'include',
                //'Cookie': cookies,
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                try {
                    AssertBag(newEntry)
                    handlers.onCreate && handlers.onCreate(newEntry)
                } catch (er) {
                    setErr("failed to decode response:")
                }
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
    }
    return (
        <NewEntryFormWrapper entryType={"bag"}>
            {/* TODO: ok?<div className={"sectionHolder"}>*/}
            <ErrorDisplay err={err}/>
            <div>{"Creating Bag: "}</div>
            {substrateBatchIn !== undefined &&
                <SubstrateBatchSelector doSelect={setSubstrateBatch} allowCreation={handlers.isTopLevel}
                                        creatorInPage={false}/>/*TODO: handle isTopLevel and creation in page*/}
            <WetnessSlider defaultValue={5} onChange={(event: Event, value: number, activeThumb: number) => {
                setWetness(value)
            }}/>
            {pcRunIn === undefined ||
                <RecentPCRunSelector doSelect={setPcRun} creatorInPage={true} allowCreation={true}/>}
            {filterSizeSelector(setFilterSize, filterSize)}
            <NewEntryNotes setNotes={setNotes}/>
            {/* Write tag area */}
            <ReaderWriterSelector txt={"Writer to write to: "} onSelect={setWriteTagTo}/>
            {/* SUBMIT AREA */}
            <input type="submit" value="Submit" className={"bottomButton"} onClick={newBagSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
        </NewEntryFormWrapper>
    )
}

export function BagImportDisplay({headerLevel, cookies}: ImportDisplayInput) { // TODO: USE
    // Required
    const [sealDate, setSealDate] = useState(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [recipe, setRecipe] = useState<SubstrateRecipeData | undefined>(undefined)
    const [filterSize, setFilterSize] = useState<string | undefined>(undefined)
    // Optional
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    //const [perms, setPerms] = useState<EntryPerms | undefined>(undefined) // TODO: use these in request
    ////const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
    const submitImportBag = () => {
        const reqd = new Map<string, any>([
            ['recipe', recipe],
            ['filterSize', filterSize],
            ['species', species]
        ])
        for (let [key, val] of reqd) {
            if (val === undefined) {
                setErr(key + " must be defined!");
                return
            }
        }
        let formData = new FormData()
        let dataObj: any = {
            sealDate: sealDate,
            recipe: recipe?._id, // MUST EXIST TODO: validate on insert
            filterSize: filterSize,
            species: species?._id, // MUST EXIST TODO: validate on insert
            //perms: perms, // TODO: validate on insert
        }

        if (subspecies !== undefined) {
            dataObj.subspecies = subspecies?._id
        }
        if (generation !== undefined) {
            dataObj.generation = generation
        }
        if (knownFruitable !== undefined) {
            dataObj.knownFruitable = knownFruitable
        }
        if (writeTagTo !== undefined) {
            dataObj.rfidWriter = writeTagTo
        }
        setFormData(formData, dataObj)
        //formData.set("data", JSON.stringify(dataObj))
        imageFile && formData.set("img", imageFile, "img")
        fetch(BaseExternalUrl + "/import/bag", {
            method: 'Post',
            body: formData,
            headers: {
                credentials: 'include',
                //'Cookie': cookies,
                'Content-type': "multipart/form-data" // TODO: auth?
                //Authorization: tokenFetch,
            },
        })
            .then(HandleTxtResponse) // TODO: make sure imports do it this way
            .then((newBagId) => {
                redirect(BaseExternalUrl + "/view/bag/" + newBagId)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"bag"}>{/* TODO: ok?<div className={"sectionHolder"}>*/}
        {/* Required Fields */}
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Seal Date: "} when={Date.now()} updateParent={setSealDate}/>
        <SubstrateRecipeSelector doSelect={setRecipe}/>{/* TODO: depth?*/}
        {filterSizeSelector(setFilterSize, filterSize)}
        <ExistingSpeciesSelector doSelect={setSpecies/*cookies={cookies}*/}/>

        {/* Optional fields*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies/*cookies={cookies}*/}/>
        <GenerationArea readonly={false} updateParent={setGeneration}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>

        <TopLevelImageSelector updateParent={setImageFile} buttonText={"Upload image"}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>

        {/* SUBMIT AREA TODO: swap out all submission buttons with a component */}
        <button onClick={submitImportBag} className={"bottomButton"}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export function BagInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<BagData>) { // TODO: DO THIS ENTIRELY!
    // TODO: do inlines need depth providers?
    const [expanded, setExpanded] = useState(expandByDefault)
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={data._id} txt={"Bag"} entryType={"bag"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <MostRecentImageDisplay data={data.mostRecentImage}/>
            <SpeciesArea readonly={true} initial={data.species}/>
            <SubspeciesArea readonly={true} initialSub={data.species} currentSpecies={data.species}
            />
            <SubstrateRecipeArea id={data.recipe} readonly={true}/>
            <SubstrateBatchArea id={data.substrateBatch} readonly={true}/>
            <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore}/>
            <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
            {data.sealDate ?
                <DateArea pre={"Sealed: "} when={data.sealDate} readonly={true}/>
                : <div></div>
            }
            <WetnessDisplay value={data.wetness}/>
            <NameArea headerTxt={"Filter Size: "} readonly={true} currentName={data.filterSize}/>
            <KnownFruitableArea initial={data.knownFruitable} readonly={true}/>
            <DisposedSaleContamArea sale={data.sale} contams={data.contamination} disposed={data.disposed}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            {/*TODO: <ProjectsArea projects={data.perms?.projectPerms.ids} readonly={true} headerLevel={headerLevel}*/}
            {/*              allowCreate={false} allowRemove={false}/>/!* TODO: ok? *!/*/}
            <div>
                <div>{"Flushes: " + (data.flushes ? data.flushes.length : 0)}</div>
            </div>
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

export function BagListPageTable({data, onClick}: ListPageItems<BagData>) {
    const cols: ListTableColumn<BagData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Species", (v)=>v.species || ""),
        NewColumn("Subspec.", (v)=>v.subspecies || ""),
    ]
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}